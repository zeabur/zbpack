package dockerfile

import (
	"log"
	"strings"

	"github.com/spf13/afero"
	"golang.org/x/text/cases"

	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new Dockerfile identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeDocker
}

func (i *identify) Match(fs afero.Fs) bool {
	fileInfo, err := afero.ReadDir(fs, ".")
	if err != nil {
		log.Println("dockerfile: read dir:", err)
		return false
	}

	converter := cases.Fold()

	for _, file := range fileInfo {
		if file.IsDir() {
			continue
		}

		foldedFilename := converter.String(file.Name())

		// We only care about {*.,}dockerfile{.*,}.
		if foldedFilename == "dockerfile" ||
			strings.HasPrefix(foldedFilename, "dockerfile.") ||
			strings.HasSuffix(foldedFilename, ".dockerfile") {
			return true
		}
	}

	return false
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	return GetMeta(options)
}

var _ plan.Identifier = (*identify)(nil)
