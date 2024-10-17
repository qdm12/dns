package setup

import (
	"net/http"
	"net/netip"

	"github.com/qdm12/dns/v2/internal/config"
	"github.com/qdm12/dns/v2/pkg/blockbuilder"
)

func BuildBlockBuilder(userSettings config.Block,
	client *http.Client,
) (blockBuilder *blockbuilder.Builder, err error) {
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

	return blockbuilder.New(settings)
}
