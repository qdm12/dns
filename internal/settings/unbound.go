package settings

import (
	"net"

	"github.com/qdm12/dns/internal/params"
	"github.com/qdm12/dns/pkg/models"
)

func getUnboundSettings(reader params.Reader) (settings models.Settings, err error) {
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
	privateAddresses, err := reader.GetPrivateAddresses()
	if err != nil {
		return settings, err
	}
	settings.BlockedIPs = append(settings.BlockedIPs, privateAddresses...)

	settings.AccessControl.Allowed = []net.IPNet{
		{
			IP:   net.IPv4zero,
			Mask: net.IPv4Mask(0, 0, 0, 0),
		},
		{
			IP:   net.IPv6zero,
			Mask: net.IPMask{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}
	return settings, nil
}
