package settings

import (
	"github.com/qdm12/dns/internal/models"
	"github.com/qdm12/dns/internal/params"
)

func GetSettings(reader params.Reader) (settings models.Settings, err error) {
	settings.Unbound, err = getUnboundSettings(reader)
	if err != nil {
		return settings, err
	}
	settings.BlockMalicious, err = reader.GetMaliciousBlocking()
	if err != nil {
		return settings, err
	}
	settings.BlockSurveillance, err = reader.GetSurveillanceBlocking()
	if err != nil {
		return settings, err
	}
	settings.BlockAds, err = reader.GetAdsBlocking()
	if err != nil {
		return settings, err
	}
	settings.CheckDNS, err = reader.GetCheckDNS()
	if err != nil {
		return settings, err
	}
	settings.UpdatePeriod, err = reader.GetUpdatePeriod()
	if err != nil {
		return settings, err
	}
	return settings, nil
}
