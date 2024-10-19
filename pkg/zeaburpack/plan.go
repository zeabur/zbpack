package zeaburpack

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

// PlanOptions is the options for Plan function.
type PlanOptions struct {
	// SubmoduleName is the of the submodule to build.
	// For example, if a directory is considered as a Go project,
	// submoduleName would be used to try file in `cmd` directory.
	// in Zeabur internal system, this is the name of the service.
	SubmoduleName *string

	// Path is the path to the project directory.
	Path *string

	// Access token for GitHub, only used when Path is a GitHub URL.
	AccessToken *string

	// AWSConfig is the AWS configuration to access S3, required if Path is an S3 URL.
	AWSConfig *plan.AWSConfig
}

// Plan returns the build plan and metadata.
func Plan(opt PlanOptions) (types.PlanType, types.PlanMeta) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if opt.Path == nil || *opt.Path == "" {
		opt.Path = &wd
	} else if !filepath.IsAbs(*opt.Path) && !strings.HasPrefix(*opt.Path, "https://") && !strings.HasPrefix(*opt.Path, "s3://") {
		p := path.Join(wd, *opt.Path)
		opt.Path = &p
	}

	var src afero.Fs
	if strings.HasPrefix(*opt.Path, "https://github.com") {
		var err error
		src, err = getGitHubSourceFromURL(*opt.Path, opt.AccessToken)
		if err != nil {
			log.Printf("unexpected github source: %v\n", err)
			return types.PlanTypeStatic, types.PlanMeta{"error": "unexpected github source", "details": err.Error()}
		}
	} else if strings.HasPrefix(*opt.Path, "s3://") {
		if opt.AWSConfig == nil {
			return types.PlanTypeStatic, types.PlanMeta{"error": "Missing AWS configuration, cannot access S3 source"}
		}

		src = getS3SourceFromURL(*opt.Path, &aws.Config{
			Region:      aws.String(opt.AWSConfig.Region),
			Credentials: credentials.NewStaticCredentials(opt.AWSConfig.AccessKeyID, opt.AWSConfig.SecretAccessKey, ""),
		})
	} else {
		src = afero.NewBasePathFs(afero.NewOsFs(), *opt.Path)
	}

	submoduleName := lo.FromPtrOr(opt.SubmoduleName, "")
	config := plan.NewProjectConfigurationFromFs(src, submoduleName)

	planner := plan.NewPlanner(
		&plan.NewPlannerOptions{
			Source:        src,
			Config:        config,
			SubmoduleName: submoduleName,
		},
		SupportedIdentifiers(config)...,
	)

	t, m := planner.Plan()
	return t, m
}

// PlanAndOutputDockerfile output dockerfile.
func PlanAndOutputDockerfile(opt PlanOptions) error {
	t, m := Plan(opt)
	dockerfile, err := generateDockerfile(
		&generateDockerfileOptions{
			planType: t,
			planMeta: m,
		},
	)
	if err != nil {
		log.Printf("Failed to generate Dockerfile: %s\n", err.Error())
		return err
	}
	println(dockerfile)
	// Remove .zeabur directory if exists
	if err := os.RemoveAll(path.Join(*opt.Path, ".zeabur")); err != nil {
		log.Printf("Failed to remove .zeabur directory: %s\n", err)
		return err
	}
	return nil
}
