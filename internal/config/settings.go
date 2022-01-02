package config

import (
	"fmt"
	"time"

	"github.com/qdm12/dns/v2/pkg/blockbuilder"
	"github.com/qdm12/dns/v2/pkg/doh"
	"github.com/qdm12/dns/v2/pkg/dot"
	"github.com/qdm12/dns/v2/pkg/filter/mapfilter"
	"github.com/qdm12/golibs/params"
)

type Settings struct {
	UpstreamType UpstreamType
	DoT          dot.ServerSettings
	DoH          doh.ServerSettings
	Cache        Cache
	Filter       mapfilter.Settings
	Metrics      Metrics
	BlockBuilder blockbuilder.Settings
	CheckDNS     bool
	Log          Log
	UpdatePeriod time.Duration
}

func (settings *Settings) get(reader *Reader) (err error) {
	reader.checkOutdatedVariables()

	settings.UpstreamType, err = getUpstreamType(reader.env)
	if err != nil {
		return err
	}

	// DNS listening address
	listeningAddress, _, err := reader.env.ListeningAddress("LISTENING_ADDRESS", params.Default(":53"))
	// Note: warning discarded since we can bind to privileged port such as 53.
	if err != nil {
		return fmt.Errorf("environment variable LISTENING_ADDRESS: %w", err)
	}
	settings.DoT.Address = listeningAddress
	settings.DoH.Address = listeningAddress

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
	settings.Log, err = getLogSettings(reader.env)
	if err != nil {
		return err
	}

	// Cache settings
	settings.Cache, err = getCacheSettings(reader)
	if err != nil {
		return err
	}

	// DoT and DoH filter settings are set later at runtime
	// using settings.FilterBuilder

	// Filter block lists building settings
	settings.BlockBuilder, err = getFilterSettings(reader)
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

	return nil
}
