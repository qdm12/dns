package settings

import (
	"github.com/qdm12/cloudflare-dns-server/internal/models"
	"github.com/qdm12/cloudflare-dns-server/internal/params"
)

func GetSettings(reader params.Reader) (settings models.Settings, err error) {
	settings.Providers, err = reader.GetProviders()
	if err != nil {
		return settings, err
	}
	settings.ListeningPort, err = reader.GetListeningPort()
	if err != nil {
		return settings, err
	}
	settings.Caching, err = reader.GetCaching()
	if err != nil {
		return settings, err
	}
	settings.IPv4, err = reader.GetIPv4()
	if err != nil {
		return settings, err
	}
	settings.IPv6, err = reader.GetIPv6()
	if err != nil {
		return settings, err
	}
	settings.VerbosityLevel, err = reader.GetVerbosity()
	if err != nil {
		return settings, err
	}
	settings.VerbosityDetailsLevel, err = reader.GetVerbosityDetails()
	if err != nil {
		return settings, err
	}
	settings.ValidationLogLevel, err = reader.GetValidationLogLevel()
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
	settings.BlockedHostnames, err = reader.GetBlockedHostnames()
	if err != nil {
		return settings, err
	}
	settings.BlockedIPs, err = reader.GetBlockedIPs()
	if err != nil {
		return settings, err
	}
	settings.AllowedHostnames, err = reader.GetUnblockedHostnames()
	if err != nil {
		return settings, err
	}
	settings.PrivateAddresses, err = reader.GetPrivateAddresses()
	if err != nil {
		return settings, err
	}
	settings.CheckUnbound, err = reader.GetCheckUnbound()
	if err != nil {
		return settings, err
	}
	settings.UpdatePeriod, err = reader.GetUpdatePeriod()
	if err != nil {
		return settings, err
	}
	return settings, nil
}
