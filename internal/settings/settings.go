package settings

import (
	"github.com/qdm12/cloudflare-dns-server/internal/models"
	"github.com/qdm12/cloudflare-dns-server/internal/params"
)

func GetSettings(reader params.Reader) (settings models.Settings, err error) {
	settings.Unbound.Providers, err = reader.GetProviders()
	if err != nil {
		return settings, err
	}
	settings.Unbound.ListeningPort, err = reader.GetListeningPort()
	if err != nil {
		return settings, err
	}
	settings.Unbound.Caching, err = reader.GetCaching()
	if err != nil {
		return settings, err
	}
	settings.Unbound.IPv4, err = reader.GetIPv4()
	if err != nil {
		return settings, err
	}
	settings.Unbound.IPv6, err = reader.GetIPv6()
	if err != nil {
		return settings, err
	}
	settings.Unbound.VerbosityLevel, err = reader.GetVerbosity()
	if err != nil {
		return settings, err
	}
	settings.Unbound.VerbosityDetailsLevel, err = reader.GetVerbosityDetails()
	if err != nil {
		return settings, err
	}
	settings.Unbound.ValidationLogLevel, err = reader.GetValidationLogLevel()
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
	settings.Unbound.BlockedHostnames, err = reader.GetBlockedHostnames()
	if err != nil {
		return settings, err
	}
	settings.Unbound.BlockedIPs, err = reader.GetBlockedIPs()
	if err != nil {
		return settings, err
	}
	settings.Unbound.AllowedHostnames, err = reader.GetUnblockedHostnames()
	if err != nil {
		return settings, err
	}
	privateAddresses, err := reader.GetPrivateAddresses()
	if err != nil {
		return settings, err
	}
	settings.Unbound.BlockedIPs = append(settings.Unbound.BlockedIPs, privateAddresses...)
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
