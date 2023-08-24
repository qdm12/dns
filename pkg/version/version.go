package version

import (
	"runtime/debug"
)

func Get() (version string) {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, dep := range buildInfo.Deps {
		if dep.Path == "github.com/qdm12/dns/v2" {
			return dep.Version
		}
	}
	return ""
}
