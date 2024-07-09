package static

import (
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cast"

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
	return utils.HasFile(fs, "index.html", "hugo.toml", "config/_default/hugo.toml", "config.toml", "mkdocs.yml")
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	if utils.HasFile(options.Source, "hugo.toml", "config/_default/hugo.toml") {
		return types.PlanMeta{"framework": "hugo"}
	}

	if utils.HasFile(options.Source, "mkdocs.yml") {
		return types.PlanMeta{"framework": "mkdocs"}
	}

	if utils.HasFile(options.Source, "config.toml") {
		config, err := utils.ReadFileToUTF8(options.Source, "config.toml")
		if err == nil && strings.Contains(string(config), "base_url") {
			ver := "0.18.0"

			if userSetVersion, err := plan.Cast(
				options.Config.Get("zola_version"), cast.ToStringE,
			).Take(); err == nil {
				ver = userSetVersion
			}

			return types.PlanMeta{"framework": "zola", "version": ver}
		}
	}

	html, err := utils.ReadFileToUTF8(options.Source, "index.html")

	if err == nil && strings.Contains(string(html), "Hexo") {
		return types.PlanMeta{"framework": "hexo"}
	}

	return types.PlanMeta{"framework": "unknown"}
}

func (i *identify) Explain(meta types.PlanMeta) []types.FieldInfo {
	fields := make([]types.FieldInfo, 0, 2)

	if framework, ok := meta["framework"]; ok {
		fields = append(fields, types.NewFrameworkFieldInfo("framework", types.PlanTypeStatic, framework))

		if framework == "zola" {
			fields = append(fields, types.FieldInfo{
				Key:         "version",
				Name:        "Zola Version",
				Description: "The Zola version for building the website.",
			})
		}
	}

	return fields
}

var _ plan.Identifier = (*identify)(nil)
