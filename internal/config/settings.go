package config

import (
	"fmt"
	"time"

	"github.com/qdm12/dns/pkg/blacklist"
	"github.com/qdm12/dns/pkg/cache"
	"github.com/qdm12/dns/pkg/doh"
	"github.com/qdm12/dns/pkg/dot"
	"github.com/qdm12/golibs/logging"
	"github.com/qdm12/golibs/params"
)

type Settings struct {
	UpstreamType UpstreamType
	DoT          dot.ServerSettings
	DoH          doh.ServerSettings
	Cache        cache.Settings
	Filter       blacklist.Settings
	Metrics      Metrics
	Blacklist    blacklist.BuilderSettings
	CheckDNS     bool
	LogLevel     logging.Level
	UpdatePeriod time.Duration
}

func (settings *Settings) get(reader *reader) (err error) {
	reader.checkOutdatedVariables()

	settings.UpstreamType, err = getUpstreamType(reader.env)
	if err != nil {
		return err
	}

	// DNS listening port
	listeningPort, err := reader.env.Port("LISTENING_PORT", params.Default("53"))
	if err != nil {
		return fmt.Errorf("environment variable LISTENING_PORT: %w", err)
	}
	settings.DoT.Port = listeningPort
	settings.DoH.Port = listeningPort

	// Metrics settings
	settings.Metrics, err = getMetricsSettings(reader)
	if err != nil {
		return err
	}

	// Resolver settings
	settings.DoT.Resolver, err = getDoTSettings(reader)
	if err != nil {
		return err
	}
	settings.DoH.Resolver, err = getDoHSettings(reader)
	if err != nil {
		return err
	}

	// Log settings
	logSettings, err := getLogSettings(reader.env)
	if err != nil {
		return err
	}
	settings.DoT.Log = logSettings
	settings.DoH.Log = logSettings

	// Cache settings
	settings.Cache, err = getCacheSettings(reader)
	if err != nil {
		return err
	}

	// DoT and DoH blacklist settings are set later at runtime
	// using settings.Blacklist

	// Blacklist building settings
	settings.Blacklist, err = getBlacklistSettings(reader)
	if err != nil {
		return err
	}

	settings.CheckDNS, err = reader.env.OnOff("CHECK_DNS", params.Default("on"))
	if err != nil {
		return fmt.Errorf("environment variable CHECK_DNS: %w", err)
	}

	settings.UpdatePeriod, err = reader.env.Duration("UPDATE_PERIOD", params.Default("24h"))
	if err != nil {
		return fmt.Errorf("environment variable UPDATE_PERIOD: %w", err)
	}

	settings.LogLevel, err = reader.env.LogLevel("LOG_LEVEL", params.Default("info"))
	if err != nil {
		return fmt.Errorf("environment variable LOG_LEVEL: %w", err)
	}

	return nil
}
