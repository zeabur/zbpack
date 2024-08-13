package source

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/afero"
)

var _ afero.Fs = &s3Fs{}

var (
	// ErrUnimplemented is returned when a method is not implemented.
	ErrUnimplemented = errors.New("unimplemented")
	// ErrReadonly is returned when this filesystem is readonly.
	ErrReadonly = errors.New("readonly")
	// ErrNotDir is returned when something is not a directory.
	ErrNotDir = errors.New("not a directory")
)

// s3Fs is a read-only filesystem abstraction for Amazon S3.
type s3Fs struct {
	S3Client *s3.S3
	Bucket   string
	Prefix   string
}

// NewS3Fs creates a new S3 filesystem with the given bucket name.
func NewS3Fs(s3Url string, sessCfg *aws.Config) afero.Fs {
	sess := session.Must(session.NewSession(sessCfg))
	client := s3.New(sess)
	bucket := strings.Split(strings.TrimPrefix(s3Url, "s3://"), "/")[0]
	prefix := strings.TrimPrefix(s3Url, "s3://"+bucket+"/")
	return &s3Fs{S3Client: client, Bucket: bucket, Prefix: prefix}
}

func (fs *s3Fs) Create(_ string) (afero.File, error) {
	return nil, ErrReadonly
}

func (fs *s3Fs) Mkdir(_ string, _ os.FileMode) error {
	return ErrReadonly
}

func (fs *s3Fs) MkdirAll(_ string, _ os.FileMode) error {
	return ErrReadonly
}

func (fs *s3Fs) Open(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

func (fs *s3Fs) OpenFile(name string, flag int, _ os.FileMode) (afero.File, error) {
	if flag&os.O_CREATE == os.O_CREATE || flag&os.O_WRONLY == os.O_WRONLY {
		return nil, ErrReadonly
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(fs.Bucket),
		Key:    aws.String(path.Join(fs.Prefix, name)),
	}

	result, err := fs.S3Client.GetObject(input)
	if err != nil {
		// if the object does not exist, maybe it's a directory
		if strings.Contains(err.Error(), "NoSuchKey") {
			listInput := &s3.ListObjectsV2Input{
				Bucket:    aws.String(fs.Bucket),
				Prefix:    aws.String(path.Join(fs.Prefix, name) + "/"),
				Delimiter: aws.String("/"),
			}
			listResult, listErr := fs.S3Client.ListObjectsV2(listInput)
			if listErr == nil && len(listResult.CommonPrefixes) > 0 {
				return &s3File{
					fs: fs,
					info: s3FileInfo{
						name:  name,
						size:  0,
						mode:  os.FileMode(0o444) | os.ModeDir,
						isDir: true,
					},
				}, nil
			}
			return nil, afero.ErrFileNotFound
		}
		return nil, fmt.Errorf("unable to read from S3: %w", err)
	}

	content, err := io.ReadAll(result.Body)

	defer func() {
		err := result.Body.Close()
		if err != nil {
			fmt.Println("error closing body", err)
		}
	}()

	if err != nil {
		return nil, fmt.Errorf("unable to read S3 object body: %w", err)
	}

	return &s3File{
		fs:     fs,
		Reader: bytes.NewReader(content),
		info:   s3FileInfo{name: name, size: *result.ContentLength},
	}, nil
}

func (fs *s3Fs) Remove(_ string) error {
	return ErrReadonly
}

func (fs *s3Fs) RemoveAll(_ string) error {
	return ErrReadonly
}

func (fs *s3Fs) Rename(_, _ string) error {
	return ErrReadonly
}

func (fs *s3Fs) Stat(name string) (os.FileInfo, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(fs.Bucket),
		Key:    aws.String(path.Join(fs.Prefix, name)),
	}

	result, err := fs.S3Client.HeadObject(input)
	if err != nil {
		return nil, fmt.Errorf("unable to stat object in S3: %w", err)
	}

	return s3FileInfo{
		name:    name,
		size:    *result.ContentLength,
		mode:    os.FileMode(0o444),
		modTime: *result.LastModified,
		isDir:   false, // S3 does not have "directories" but emulates them using keys
	}, nil
}

func (fs *s3Fs) Name() string {
	return "S3Fs"
}

func (fs *s3Fs) Chmod(_ string, _ os.FileMode) error {
	return ErrReadonly
}

func (fs *s3Fs) Chtimes(_ string, _ time.Time, _ time.Time) error {
	return ErrReadonly
}

func (fs *s3Fs) Chown(_ string, _, _ int) error {
	return ErrReadonly
}

var _ afero.File = &s3File{}

type s3File struct {
	fs *s3Fs
	*bytes.Reader
	info os.FileInfo
}

func (f *s3File) Close() error {
	return nil
}

func (f *s3File) Name() string {
	return f.info.Name()
}

// Readdir reads the contents of the directory associated with file and returns
// a slice of up to n FileInfo values, as would be returned by Lstat, in directory order.
// Subsequent calls on the same file will yield further FileInfos.
func (f *s3File) Readdir(count int) ([]os.FileInfo, error) {
	if !f.info.IsDir() {
		return nil, ErrNotDir
	}

	prefix := path.Join(f.fs.Prefix, f.Name())
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(f.fs.Bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	}

	result, err := f.fs.S3Client.ListObjectsV2(input)
	if err != nil {
		return nil, fmt.Errorf("unable to list objects in S3: %w", err)
	}

	var fileInfos []os.FileInfo
	for _, obj := range result.Contents {
		fileInfos = append(fileInfos, s3FileInfo{
			name:    strings.TrimPrefix(*obj.Key, prefix),
			size:    *obj.Size,
			mode:    os.FileMode(0o444),
			modTime: *obj.LastModified,
			isDir:   false,
		})
	}

	for _, p := range result.CommonPrefixes {
		dirName := strings.TrimSuffix(strings.TrimPrefix(*p.Prefix, prefix), "/")
		fileInfos = append(fileInfos, s3FileInfo{
			name:  dirName,
			size:  0,
			mode:  os.FileMode(0o444) | os.ModeDir,
			isDir: true,
		})
	}
	if count > 0 && len(fileInfos) > count {
		fileInfos = fileInfos[:count]
	}

	return fileInfos, nil
}

// Readdirnames reads and returns a slice of names from the directory f.
func (f *s3File) Readdirnames(n int) ([]string, error) {
	fis, err := f.Readdir(n)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(fis))
	for _, fi := range fis {
		names = append(names, fi.Name())
	}
	return names, nil
}

func (f *s3File) Stat() (os.FileInfo, error) {
	return f.info, nil
}

func (f *s3File) Sync() error {
	return ErrReadonly
}

func (f *s3File) Truncate(_ int64) error {
	return ErrReadonly
}

func (f *s3File) Write(_ []byte) (n int, err error) {
	return 0, ErrReadonly
}

func (f *s3File) WriteAt(_ []byte, _ int64) (n int, err error) {
	return 0, ErrReadonly
}

func (f *s3File) WriteString(_ string) (ret int, err error) {
	return 0, ErrReadonly
}

type s3FileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (fi s3FileInfo) Name() string       { return fi.name }
func (fi s3FileInfo) Size() int64        { return fi.size }
func (fi s3FileInfo) Mode() os.FileMode  { return fi.mode }
func (fi s3FileInfo) ModTime() time.Time { return fi.modTime }
func (fi s3FileInfo) IsDir() bool        { return fi.isDir }
func (fi s3FileInfo) Sys() interface{}   { return nil }
