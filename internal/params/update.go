package params

import (
	"time"

	libparams "github.com/qdm12/golibs/params"
)

// GetUpdatePeriod obtains the period to use to update the block lists and cryptographic files
// and restart Unbound from the environment variable DNS_UPDATE_PERIOD
func (p *reader) GetUpdatePeriod() (period time.Duration, err error) {
	s, err := p.envParams.GetEnv("UPDATE_PERIOD", libparams.Default("24h"))
	if err != nil {
		return period, err
	}
	return time.ParseDuration(s)
}
