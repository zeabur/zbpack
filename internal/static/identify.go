package static

import (
	"strconv"
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
	serverless := utils.GetExplicitServerlessConfig(options.Config).TakeOr(true)

	planMeta := types.PlanMeta{"serverless": strconv.FormatBool(serverless)}

	if utils.HasFile(options.Source, "hugo.toml", "config/_default/hugo.toml") {
		planMeta["framework"] = "hugo"
		return planMeta
	}

	if utils.HasFile(options.Source, "mkdocs.yml") {
		planMeta["framework"] = "mkdocs"
		return planMeta
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

			planMeta["framework"] = "zola"
			planMeta["version"] = ver
			return planMeta
		}
	}

	html, err := utils.ReadFileToUTF8(options.Source, "index.html")

	if err == nil && strings.Contains(string(html), "Hexo") {
		planMeta["framework"] = "hexo"
		return planMeta
	}

	return planMeta
}

var _ plan.Identifier = (*identify)(nil)
