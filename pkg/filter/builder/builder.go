package builder

import (
	"context"
	"net/http"

	"inet.af/netaddr"
)

type Interface interface {
	All(ctx context.Context, settings Settings) (
		blockedHostnames []string, blockedIPs []netaddr.IP,
		blockedIPPrefixes []netaddr.IPPrefix, errs []error)
	Hostnames(ctx context.Context,
		blockMalicious, blockAds, blockSurveillance bool,
		additionalBlockedHostnames, allowedHostnames []string) (
		blockedHostnames []string, errs []error)
	IPs(ctx context.Context,
		blockMalicious, blockAds, blockSurveillance bool,
		allowedIPs, additionalBlockedIPs []netaddr.IP,
		allowedIPPrefixes, additionalBlockedIPPrefixes []netaddr.IPPrefix) (
		blockedIPs []netaddr.IP, blockedIPPrefixes []netaddr.IPPrefix, errs []error)
}

func New(client *http.Client) Interface {
	return &builder{
		client: client,
		// TODO cache blocked IPs and hostnames after first request?
	}
}

type builder struct {
	client *http.Client
}
