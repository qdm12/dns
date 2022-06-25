package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/qdm12/dns/v2/internal/config/settings"
)

type Reader struct {
	warner Warner
}

type Warner interface {
	Warn(s string)
}

func New(warner Warner) *Reader {
	return &Reader{
		warner: warner,
	}
}

func (r *Reader) Read() (settings settings.Settings, err error) { //nolint:cyclop
	warnings := checkOutdatedVariables()
	for _, warning := range warnings {
		r.warner.Warn(warning)
	}

	settings.Upstream = strings.ToLower(os.Getenv("UPSTREAM_TYPE"))
	settings.ListeningAddress = os.Getenv("LISTENING_ADDRESS")

	settings.Block, err = readBlock()
	if err != nil {
		return settings, fmt.Errorf("cannot read block settings: %w", err)
	}

	settings.Cache, err = readCache()
	if err != nil {
		return settings, fmt.Errorf("cannot read cache settings: %w", err)
	}

	settings.DoH, err = readDoH()
	if err != nil {
		return settings, fmt.Errorf("cannot read DoH settings: %w", err)
	}

	settings.DoT, err = readDoT()
	if err != nil {
		return settings, fmt.Errorf("cannot read DoT settings: %w", err)
	}

	settings.Log, err = readLog()
	if err != nil {
		return settings, fmt.Errorf("cannot read log settings: %w", err)
	}

	settings.MiddlewareLog, err = readMiddlewareLog()
	if err != nil {
		return settings, fmt.Errorf("cannot read middleware log settings: %w", err)
	}

	settings.Metrics, err = readMetrics()
	if err != nil {
		return settings, fmt.Errorf("cannot read metrics settings: %w", err)
	}

	settings.CheckDNS, err = envToBoolPtr("CHECK_DNS")
	if err != nil {
		return settings, fmt.Errorf("environment variable CHECK_DNS: %w", err)
	}

	settings.UpdatePeriod, err = envToDurationPtr("UPDATE_PERIOD")
	if err != nil {
		return settings, fmt.Errorf("environment variable UPDATE_PERIOD: %w", err)
	}

	return settings, nil
}
