package env

import (
	"fmt"
	"os"

	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/gosettings/sources/env"
)

type Reader struct {
	env    env.Env
	warner Warner
}

type Warner interface {
	Warn(s string)
}

func New(warner Warner) *Reader {
	return &Reader{
		warner: warner,
		env:    *env.New(os.Environ(), nil),
	}
}

func (r *Reader) Read() (settings settings.Settings, err error) {
	warnings := checkOutdatedVariables()
	for _, warning := range warnings {
		r.warner.Warn(warning)
	}

	settings.Upstream = r.env.String("UPSTREAM_TYPE")
	settings.ListeningAddress = r.env.Get("LISTENING_ADDRESS")

	settings.Block, err = r.readBlock()
	if err != nil {
		return settings, fmt.Errorf("block settings: %w", err)
	}

	settings.Cache, err = r.readCache()
	if err != nil {
		return settings, fmt.Errorf("cache settings: %w", err)
	}

	settings.DoH, err = r.readDoH()
	if err != nil {
		return settings, fmt.Errorf("DoH settings: %w", err)
	}

	settings.DoT, err = r.readDoT()
	if err != nil {
		return settings, fmt.Errorf("DoT settings: %w", err)
	}

	settings.Log, err = r.readLog()
	if err != nil {
		return settings, fmt.Errorf("log settings: %w", err)
	}

	settings.MiddlewareLog, err = r.readMiddlewareLog()
	if err != nil {
		return settings, fmt.Errorf("middleware log settings: %w", err)
	}

	settings.Metrics = r.readMetrics()

	settings.CheckDNS, err = r.env.BoolPtr("CHECK_DNS")
	if err != nil {
		return settings, fmt.Errorf("environment variable CHECK_DNS: %w", err)
	}

	settings.UpdatePeriod, err = r.env.DurationPtr("UPDATE_PERIOD")
	if err != nil {
		return settings, fmt.Errorf("environment variable UPDATE_PERIOD: %w", err)
	}

	return settings, nil
}
