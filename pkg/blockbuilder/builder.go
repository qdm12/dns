package blockbuilder

import (
	"fmt"
	"net/http"
	"net/netip"
)

func New(settings Settings) (builder *Builder, err error) {
	settings.SetDefaults()

	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("settings validation: %w", err)
	}

	return &Builder{
		client:               settings.Client,
		blockMalicious:       *settings.BlockMalicious,
		blockAds:             *settings.BlockAds,
		blockSurveillance:    *settings.BlockSurveillance,
		allowedHosts:         settings.AllowedHosts,
		allowedIPs:           settings.AllowedIPs,
		allowedIPPrefixes:    settings.AllowedIPPrefixes,
		addBlockedHosts:      settings.AddBlockedHosts,
		addBlockedIPs:        settings.AddBlockedIPs,
		addBlockedIPPrefixes: settings.AddBlockedIPPrefixes,
		// TODO cache blocked IPs and hostnames after first request?
	}, nil
}

type Builder struct {
	client               *http.Client
	blockMalicious       bool
	blockAds             bool
	blockSurveillance    bool
	allowedHosts         []string
	allowedIPs           []netip.Addr
	allowedIPPrefixes    []netip.Prefix
	addBlockedHosts      []string
	addBlockedIPs        []netip.Addr
	addBlockedIPPrefixes []netip.Prefix
}
