package params

import (
	"github.com/qdm12/golibs/params"
)

func (r *reader) GetVersion() string {
	version, _ := r.envParams.GetEnv("VERSION", params.Default("?"))
	return version
}

func (r *reader) GetBuildDate() string {
	buildDate, _ := r.envParams.GetEnv("BUILD_DATE", params.Default("?"))
	return buildDate
}

func (r *reader) GetVcsRef() string {
	buildDate, _ := r.envParams.GetEnv("VCS_REF", params.Default("?"))
	return buildDate
}
