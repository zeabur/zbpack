package python

import (
	"github.com/zeabur/zbpack/internal/utils"
)

const defaultPython3Version = "3.11"

func getPython3Version(versionRange string) string {
	return utils.ConstraintToVersion(versionRange, defaultPython3Version)
}
