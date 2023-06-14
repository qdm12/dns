package setup

import (
	"net/http"
	"net/netip"

	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/dns/v2/pkg/blockbuilder"
)

func BuildBlockBuilder(userSettings settings.Block,
	client *http.Client) (blockBuilder *blockbuilder.Builder) {
	settings := blockbuilder.Settings{
		Client:            client,
		BlockMalicious:    userSettings.BlockMalicious,
		BlockAds:          userSettings.BlockAds,
		BlockSurveillance: userSettings.BlockSurveillance,
	}

	settings.AllowedHosts = make([]string, len(userSettings.AllowedHosts))
	copy(settings.AllowedHosts, userSettings.AllowedHosts)

	settings.AllowedIPs = make([]netip.Addr, len(userSettings.AllowedIPs))
	copy(settings.AllowedIPs, userSettings.AllowedIPs)

	settings.AllowedIPPrefixes = make([]netip.Prefix, len(userSettings.AllowedIPPrefixes))
	copy(settings.AllowedIPPrefixes, userSettings.AllowedIPPrefixes)

	settings.AddBlockedHosts = make([]string, len(userSettings.AddBlockedHosts))
	copy(settings.AddBlockedHosts, userSettings.AddBlockedHosts)

	settings.AddBlockedIPs = make([]netip.Addr, len(userSettings.AddBlockedIPs))
	copy(settings.AddBlockedIPs, userSettings.AddBlockedIPs)

	settings.AddBlockedIPPrefixes = make([]netip.Prefix, len(userSettings.AddBlockedIPPrefixes))
	copy(settings.AddBlockedIPPrefixes, userSettings.AddBlockedIPPrefixes)
	if *userSettings.RebindingProtection {
		privateIPPrefixes, err := getPrivateIPPrefixes()
		if err != nil {
			panic(err)
		}
		settings.AddBlockedIPPrefixes = append(settings.AddBlockedIPPrefixes, privateIPPrefixes...)
	}

	return blockbuilder.New(settings)
}

func getPrivateIPPrefixes() (privateIPPrefixes []netip.Prefix, err error) {
	privateCIDRs := []string{
		// IPv4 private addresses
		"127.0.0.1/8",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
		// IPv6 private addresses
		"::1/128",
		"fc00::/7",
		"fe80::/10",
		// Private IPv4 addresses wrapped in IPv6
		"::ffff:7f00:1/104", // 127.0.0.1/8
		"::ffff:a00:0/104",  // 10.0.0.0/8
		"::ffff:ac10:0/108", // 172.16.0.0/12
		"::ffff:c0a8:0/112", // 192.168.0.0/16
		"::ffff:a9fe:0/112", // 169.254.0.0/16
	}
	privateIPPrefixes = make([]netip.Prefix, len(privateCIDRs))
	for i := range privateCIDRs {
		privateIPPrefixes[i], err = netip.ParsePrefix(privateCIDRs[i])
		if err != nil {
			return nil, err
		}
	}

	return privateIPPrefixes, nil
}
