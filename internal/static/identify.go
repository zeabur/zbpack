package static

import (
	"strings"

	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

// static site generator (hugo, gatsby, etc) detection
type identify struct{}

// NewIdentifier returns a new static site generator identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeStatic
}

func (i *identify) Match(fs afero.Fs) bool {
	return utils.HasFile(fs, "index.html", "hugo.toml", "config/_default/hugo.toml")
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {

	if utils.HasFile(options.Source, "hugo.toml", "config/_default/hugo.toml") {
		return types.PlanMeta{"framework": "hugo"}
	}

	html, err := afero.ReadFile(options.Source, "index.html")

	if err == nil && strings.Contains(string(html), "Hugo") {
		return types.PlanMeta{"framework": "hugo"}
	}

	if err == nil && strings.Contains(string(html), "Hexo") {
		return types.PlanMeta{"framework": "hexo"}
	}

	return types.PlanMeta{"framework": "unknown"}
}

var _ plan.Identifier = (*identify)(nil)
