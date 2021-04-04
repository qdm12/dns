package config

import (
	"net"

	"github.com/qdm12/dns/pkg/unbound"
	"github.com/qdm12/golibs/params"
)

func getUnboundSettings(reader *reader) (settings unbound.Settings, err error) {
	settings.Providers, err = getProviders(reader)
	if err != nil {
		return settings, err
	}
	settings.ListeningPort, err = reader.env.Port("LISTENINGPORT", params.Default("53"))
	if err != nil {
		return settings, err
	}
	settings.Caching, err = reader.env.OnOff("CACHING", params.Default("off"))
	if err != nil {
		return settings, err
	}
	settings.IPv4, err = reader.env.OnOff("IPV4", params.Default("on"))
	if err != nil {
		return settings, err
	}
	settings.IPv6, err = reader.env.OnOff("IPV6", params.Default("off"))
	if err != nil {
		return settings, err
	}

	verbosityDetails, err := reader.env.IntRange("VERBOSITY", 0, 5, params.Default("1"))
	if err != nil {
		return settings, err
	}
	settings.VerbosityLevel = uint8(verbosityDetails)

	verbosityDetailsLevel, err := reader.env.IntRange("VERBOSITY_DETAILS", 0, 4, params.Default("0"))
	if err != nil {
		return settings, err
	}
	settings.VerbosityDetailsLevel = uint8(verbosityDetailsLevel)

	validationLogLevel, err := reader.env.IntRange("VALIDATION_LOGLEVEL", 0, 2, params.Default("0"))
	if err != nil {
		return settings, err
	}
	settings.ValidationLogLevel = uint8(validationLogLevel)

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
