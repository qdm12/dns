package filter

import (
	"context"
	"net/http"

	"inet.af/netaddr"
)

type Builder interface {
	All(ctx context.Context, settings BuilderSettings) (
		blockedHostnames []string, blockedIPs []netaddr.IP,
		blockedIPPrefixes []netaddr.IPPrefix, errs []error)
	Hostnames(ctx context.Context,
		blockMalicious, blockAds, blockSurveillance bool,
		additionalBlockedHostnames, allowedHostnames []string) (
		blockedHostnames []string, errs []error)
	IPs(ctx context.Context,
		blockMalicious, blockAds, blockSurveillance bool,
		additionalBlockedIPs []netaddr.IP, additionalBlockedIPPrefixes []netaddr.IPPrefix) (
		blockedIPs []netaddr.IP, blockedIPPrefixes []netaddr.IPPrefix, errs []error)
}

func NewBuilder(client *http.Client) Builder {
	return &builder{
		client: client,
		// TODO cache blocked IPs and hostnames after first request?
	}
}

type builder struct {
	client *http.Client
}
