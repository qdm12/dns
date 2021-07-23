package config

import (
	"fmt"
	"time"

	"github.com/qdm12/dns/pkg/blacklist"
	"github.com/qdm12/dns/pkg/unbound"
	"github.com/qdm12/golibs/params"
)

type Settings struct {
	Unbound      unbound.Settings
	Blacklist    blacklist.BuilderSettings
	CheckDNS     bool
	UpdatePeriod time.Duration
}

func (settings *Settings) get(reader *reader) (err error) {
	settings.Unbound, err = getUnboundSettings(reader)
	if err != nil {
		return err
	}

	// Blacklist building settings
	settings.Blacklist, err = getBlacklistSettings(reader)
	if err != nil {
		return err
	}
	settings.CheckDNS, err = reader.env.OnOff("CHECK_DNS", params.Default("on"),
		params.RetroKeys([]string{"CHECK_UNBOUND"}, reader.onRetroActive))
	if err != nil {
		return fmt.Errorf("environment variable CHECK_DNS: %w", err)
	}
	settings.UpdatePeriod, err = reader.env.Duration("UPDATE_PERIOD", params.Default("24h"))
	if err != nil {
		return fmt.Errorf("environment variable UPDATE_PERIOD: %w", err)
	}

	return nil
}
