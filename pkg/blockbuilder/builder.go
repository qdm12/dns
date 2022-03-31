package blockbuilder

import (
	"context"
	"net/http"

	"inet.af/netaddr"
)

var _ Interface = (*Builder)(nil)

type Interface interface {
	BuildAll(ctx context.Context) Result
}

func New(settings Settings) *Builder {
	settings.SetDefaults()

	return &Builder{
		client:               settings.Client,
		blockMalicious:       settings.BlockMalicious,
		blockAds:             settings.BlockAds,
		blockSurveillance:    settings.BlockSurveillance,
		allowedHosts:         settings.AllowedHosts,
		allowedIPs:           settings.AllowedIPs,
		allowedIPPrefixes:    settings.AllowedIPPrefixes,
		addBlockedHosts:      settings.AddBlockedHosts,
		addBlockedIPs:        settings.AddBlockedIPs,
		addBlockedIPPrefixes: settings.AddBlockedIPPrefixes,
		// TODO cache blocked IPs and hostnames after first request?
	}
}

type Builder struct {
	client               *http.Client
	blockMalicious       bool
	blockAds             bool
	blockSurveillance    bool
	allowedHosts         []string
	allowedIPs           []netaddr.IP
	allowedIPPrefixes    []netaddr.IPPrefix
	addBlockedHosts      []string
	addBlockedIPs        []netaddr.IP
	addBlockedIPPrefixes []netaddr.IPPrefix
}
