package python

import (
	"strconv"

	"github.com/zeabur/zbpack/pkg/types"
)

// StaticFlag is the static flag of a Python project.
type StaticFlag uint64

const (
	// StaticModeDisabled indicates that no static files should be
	// generated or hosted.
	StaticModeDisabled StaticFlag = 0

	// StaticModeDjango indicates that we need to prepare the
	// static assets for a Django project.
	StaticModeDjango StaticFlag = 1 << (iota - 1)

	// StaticModeNginx indicates that we need to host the static
	// files with Nginx. The Python or WSGI server must be listened
	// on 8000 for reverse proxying by Nginx. The "8000" is
	// configured by our nginx.conf in `python.go`.
	StaticModeNginx StaticFlag = 1 << (iota - 1)
)

// StaticInfo is the static info of a Python project.
type StaticInfo struct {
	// Flag indicates the static flag of a Python project.
	Flag StaticFlag

	// StaticURLPath indicates where the static files can be
	// accessed from the web. For example, `/static` means that
	// the static files can be accessed from `http://example.com/static`.
	StaticURLPath string

	// StaticHostDir indicates where the static files are located
	// in the host. For example, `/app/static` means that the static
	// files are located at `/app/static` in the container.
	StaticHostDir string
}

// Enabled returns true if the static files should be hosted with Nginx.
func (i StaticInfo) Enabled() bool {
	return i.Flag != StaticModeDisabled
}

// DjangoEnabled returns true if the static files should be prepared
// for a Django project.
func (i StaticInfo) DjangoEnabled() bool {
	return i.Flag&StaticModeDjango != 0
}

// NginxEnabled returns true if the static files should be hosted
// with Nginx.
func (i StaticInfo) NginxEnabled() bool {
	return i.Flag&StaticModeNginx != 0
}

// Meta turns this structure into a partial PlanMeta.
func (i StaticInfo) Meta() types.PlanMeta {
	if !i.Enabled() {
		return nil
	}

	return types.PlanMeta{
		"static-flag":     strconv.FormatUint(uint64(i.Flag), 16),
		"static-url-path": i.StaticURLPath,
		"static-host-dir": i.StaticHostDir,
	}
}

// staticInfoFromMeta creates a StaticInfo from a partial PlanMeta.
func staticInfoFromMeta(meta types.PlanMeta) StaticInfo {
	info := StaticInfo{}

	if flag, ok := meta["static-flag"]; ok {
		if f, err := strconv.ParseUint(flag, 16, 64); err == nil {
			info.Flag = StaticFlag(f)
		}
	}

	if path, ok := meta["static-url-path"]; ok {
		info.StaticURLPath = path
	}

	if path, ok := meta["static-host-dir"]; ok {
		info.StaticHostDir = path
	}

	return info
}
