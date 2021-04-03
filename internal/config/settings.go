package config

import (
	"time"

	"github.com/qdm12/dns/pkg/unbound"
	"github.com/qdm12/golibs/params"
)

type Settings struct {
	Unbound           unbound.Settings
	BlockMalicious    bool
	BlockAds          bool
	BlockSurveillance bool
	CheckDNS          bool
	UpdatePeriod      time.Duration
}

func (settings *Settings) get(reader *reader) (err error) {
	settings.Unbound, err = getUnboundSettings(reader)
	if err != nil {
		return err
	}
	settings.BlockMalicious, err = reader.env.OnOff("BLOCK_MALICIOUS", params.Default("on"))
	if err != nil {
		return err
	}
	settings.BlockSurveillance, err = reader.env.OnOff("BLOCK_SURVEILLANCE", params.Default("off"),
		params.RetroKeys([]string{"BLOCK_NSA"}, reader.onRetroActive))
	if err != nil {
		return err
	}
	settings.BlockAds, err = reader.env.OnOff("BLOCK_ADS", params.Default("off"))
	if err != nil {
		return err
	}
	settings.CheckDNS, err = reader.env.OnOff("CHECK_DNS", params.Default("on"),
		params.RetroKeys([]string{"CHECK_UNBOUND"}, reader.onRetroActive))
	if err != nil {
		return err
	}
	settings.UpdatePeriod, err = reader.env.Duration("UPDATE_PERIOD", params.Default("24h"))
	if err != nil {
		return err
	}

	return nil
}
