package config

import (
	"fmt"

	"github.com/qdm12/dns/pkg/unbound"
	"github.com/qdm12/golibs/params"
	"inet.af/netaddr"
)

func getUnboundSettings(reader *reader) (settings unbound.Settings, err error) {
	settings.Providers, err = getProviders(reader)
	if err != nil {
		return settings, err
	}
	settings.ListeningPort, err = reader.env.Port("LISTENINGPORT", params.Default("53"))
	if err != nil {
		return settings, fmt.Errorf("environment variable LISTENINGPORT: %w", err)
	}
	settings.Caching, err = reader.env.OnOff("CACHING", params.Default("off"))
	if err != nil {
		return settings, fmt.Errorf("environment variable CACHING: %w", err)
	}
	settings.IPv4, err = reader.env.OnOff("IPV4", params.Default("on"))
	if err != nil {
		return settings, fmt.Errorf("environment variable IPV4: %w", err)
	}
	settings.IPv6, err = reader.env.OnOff("IPV6", params.Default("off"))
	if err != nil {
		return settings, fmt.Errorf("environment variable IPV6: %w", err)
	}

	verbosityDetails, err := reader.env.IntRange("VERBOSITY", 0, 5, params.Default("1")) //nolint:gomnd
	if err != nil {
		return settings, fmt.Errorf("environment variable VERBOSITY: %w", err)
	}
	settings.VerbosityLevel = uint8(verbosityDetails)

	verbosityDetailsLevel, err := reader.env.IntRange("VERBOSITY_DETAILS", 0, 4, params.Default("0")) //nolint:gomnd
	if err != nil {
		return settings, fmt.Errorf("environment variable VERBOSITY_DETAILS: %w", err)
	}
	settings.VerbosityDetailsLevel = uint8(verbosityDetailsLevel)

	validationLogLevel, err := reader.env.IntRange("VALIDATION_LOGLEVEL", 0, 2, params.Default("0")) //nolint:gomnd
	if err != nil {
		return settings, fmt.Errorf("environment variable VALIDATION_LOGLEVEL: %w", err)
	}
	settings.ValidationLogLevel = uint8(validationLogLevel)

	settings.AccessControl.Allowed = []netaddr.IPPrefix{
		{IP: netaddr.IPv4(0, 0, 0, 0)},
		{IP: netaddr.IPv6Raw([16]byte{})},
	}
	return settings, nil
}
