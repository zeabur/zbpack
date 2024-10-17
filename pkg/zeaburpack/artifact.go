package zeaburpack

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/codeclysm/extract/v3"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/frontend/dockerfile/builder"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/tonistiigi/fsutil"
	"github.com/zeabur/zbpack/internal/static"
	zbserverless "github.com/zeabur/zbpack/pkg/zeaburpack/serverless"
	"golang.org/x/sync/errgroup"

	// allow using "docker-container://" protocol in the buildkit client.
	_ "github.com/moby/buildkit/client/connhelper/dockercontainer"
)

// ImageBuilder is a builder for building images.
type ImageBuilder struct {
	// Path is the path of this project.
	Path string

	// PlanMeta is the planned metadata of this project.
	PlanMeta map[string]string

	// ResultImage is the name of the image that will be built.
	ResultImage string

	// DockerfileContent is the content of the Dockerfile to create.
	DockerfileContent string

	// Stage indicates the stage of the image to pick.
	// Empty = default layer.
	Stage string

	// BuildArgs are the build arguments to pass to the Docker build.
	BuildArgs map[string]string

	// LogWriter is the writer to write the buildkit logs to.
	// If not provided, the logs will be written to os.Stderr.
	LogWriter io.Writer
}

// Artifact is the output of the image build.
type Artifact struct {
	// dockerTar is the tar file of the Docker image.
	dockerTar *string

	// dotZeaburDirectory is the directory of the Zeabur serverless artifact.
	dotZeaburDirectory *string
}

// GetDockerTar returns the Docker tar file path of the artifact.
func (a Artifact) GetDockerTar() (string, bool) {
	if a.dockerTar == nil {
		return "", false
	}

	return *a.dockerTar, true
}

// GetDotZeaburDirectory returns the Zeabur serverless artifact directory.
func (a Artifact) GetDotZeaburDirectory() (string, bool) {
	if a.dotZeaburDirectory == nil {
		return "", false
	}

	return *a.dotZeaburDirectory, true
}

// BuildImage builds the image, transforms it to the expected format,
// and returns the artifact.
func (b *ImageBuilder) BuildImage(ctx context.Context) (*Artifact, error) {
	imageType, attrs := GetImageType(b.DockerfileContent, b.Stage)

	dockerfileDir, err := os.MkdirTemp("", "zbpack-image-")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(dockerfileDir)
	}()

	err = os.WriteFile(path.Join(dockerfileDir, "Dockerfile"), []byte(b.DockerfileContent), 0o644)
	if err != nil {
		return nil, fmt.Errorf("write Dockerfile: %w", err)
	}

	archivePath, err := b.buildImageToArchive(ctx, imageType, dockerfileDir)
	if err != nil {
		return nil, fmt.Errorf("build image to archive: %w", err)
	}

	switch imageType {
	case "static":
		dir, err := extractTarToDirectory(ctx, archivePath)
		if err != nil {
			return nil, fmt.Errorf("extract tar to directory: %w", err)
		}
		defer func() {
			_ = os.RemoveAll(dir)
		}()

		dotZeaburDirectory := b.dotZeaburDirectory()

		err = static.TransformServerless(dir, dotZeaburDirectory, b.PlanMeta)
		if err != nil {
			return nil, fmt.Errorf("transform static: %w", err)
		}

		return &Artifact{
			dotZeaburDirectory: &dotZeaburDirectory,
		}, nil
	case "serverless":
		dir, err := extractTarToDirectory(ctx, archivePath)
		if err != nil {
			return nil, fmt.Errorf("extract tar to directory: %w", err)
		}

		if attrs["serverless-transformer"] == "" {
			return nil, errors.New("serverless transformer is not set")
		}

		transformer, ok := zbserverless.GetTransformer(attrs["serverless-transformer"])
		if !ok {
			return nil, fmt.Errorf("no such serverless transformer: %s", attrs["serverless-transformer"])
		}

		dotZeaburDirectory := b.dotZeaburDirectory()

		err = transformer(dir, dotZeaburDirectory, b.PlanMeta)
		if err != nil {
			return nil, fmt.Errorf("transform serverless: %w", err)
		}

		return &Artifact{
			dotZeaburDirectory: &dotZeaburDirectory,
		}, nil
	case "containerized":
		dir, err := os.MkdirTemp("", "zbpack-docker-tar-")
		if err != nil {
			return nil, fmt.Errorf("create temp file: %w", err)
		}

		newArtifactTar := path.Join(dir, "artifact.tar")

		err = os.Rename(archivePath, newArtifactTar)
		if err != nil {
			return nil, fmt.Errorf("rename artifact tar: %w", err)
		}

		return &Artifact{
			dockerTar: &newArtifactTar,
		}, nil
	}

	return nil, fmt.Errorf("unknown image type: %s", imageType)
}

// buildImageToArchive builds the image to an archive,
// returning the path to the archive.
//
// For serverless and static images, the image is exported to a tar file.
// For containerized images, the image is exported to a Docker tar file
// that can be copied with `docker load`.
func (b *ImageBuilder) buildImageToArchive(ctx context.Context, imageType string, dockerfileDir string) (artifactPath string, err error) {
	contextFS, err := fsutil.NewFS(b.Path)
	if err != nil {
		return "", fmt.Errorf("open context directory: %w", err)
	}

	dockerfileFS, err := fsutil.NewFS(dockerfileDir)
	if err != nil {
		return "", fmt.Errorf("open dockerfile directory: %w", err)
	}

	exporter := client.ExporterDocker
	switch imageType {
	case "serverless", "static":
		exporter = client.ExporterTar
	case "containerized":
		exporter = client.ExporterDocker
	}

	artifactTarPath := path.Join(dockerfileDir, "artifact.tar")
	artifact, err := os.OpenFile(artifactTarPath, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("open artifact tar to write: %w", err)
	}
	defer func() {
		_ = artifact.Close()
	}()

	buildkitHost := os.Getenv("BUILDKIT_HOST")
	if buildkitHost == "" {
		return "", errors.New("BUILDKIT_HOST is not set")
	}

	c, err := client.New(ctx, buildkitHost)
	if err != nil {
		return "", fmt.Errorf("create buildkit client: %w", err)
	}

	frontendAttrs := map[string]string{
		"filename": "Dockerfile",
		"platform": fmt.Sprintf("linux/%s", runtime.GOARCH),
		"target":   b.Stage,
	}
	for key, value := range b.BuildArgs {
		frontendAttrs["build-arg:"+key] = value
	}

	solveOpt := client.SolveOpt{
		Exports: []client.ExportEntry{
			{
				Type: exporter,
				Attrs: map[string]string{
					"name": b.ResultImage,
				},
				Output: func(_ map[string]string) (io.WriteCloser, error) {
					return artifact, nil
				},
			},
		},
		LocalMounts: map[string]fsutil.FS{
			"context":    contextFS,
			"dockerfile": dockerfileFS,
		},
		Frontend:      "dockerfile.v0",
		FrontendAttrs: frontendAttrs,
	}

	ch := make(chan *client.SolveStatus)
	eg, ectx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		_, err := c.Build(ectx, solveOpt, "zbpack-buildkit-docker", builder.Build, ch)
		return err
	})
	eg.Go(func() error {
		d, err := progressui.NewDisplay(b.getLogWriter(), progressui.AutoMode)
		if err != nil {
			return err
		}
		_, err = d.UpdateFrom(context.TODO(), ch)
		return err
	})

	if err := eg.Wait(); err != nil {
		return "", err
	}

	return artifactTarPath, nil
}

func (b *ImageBuilder) dotZeaburDirectory() string {
	d := path.Join(b.Path, ".zeabur")
	if _, err := os.Stat(d); err == nil {
		_ = os.RemoveAll(d)
	}

	_ = os.Mkdir(d, 0o755)

	return d
}

func (b *ImageBuilder) getLogWriter() io.Writer {
	if b.LogWriter == nil {
		return os.Stderr
	}

	return b.LogWriter
}

func extractTarToDirectory(ctx context.Context, tarPath string) (string, error) {
	dir, err := os.MkdirTemp("", "zbpack-tar-rootfs-")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}

	tarFile, err := os.Open(tarPath)
	if err != nil {
		return "", fmt.Errorf("open tar file: %w", err)
	}
	defer func() {
		if err := tarFile.Close(); err != nil {
			log.Println("close tar file:", err)
		}
	}()

	err = extract.Tar(ctx, tarFile, dir, nil)
	if err != nil {
		return "", fmt.Errorf("extract tar: %w", err)
	}

	return dir, nil
}
