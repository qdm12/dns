package blacklist

import (
	"context"
	"net"
	"net/http"
)

type Builder interface {
	All(ctx context.Context,
		blockMalicious, blockAds, blockSurveillance bool,
		additionalBlockedHostnames, allowedHostnames []string,
		additionalBlockedIPs []net.IP, additionalBlockedIPNets []*net.IPNet) (
		blockedHostnames []string, blockedIPs []net.IP,
		blockedIPNets []*net.IPNet, errs []error)
	Hostnames(ctx context.Context,
		blockMalicious, blockAds, blockSurveillance bool,
		additionalBlockedHostnames, allowedHostnames []string) (
		blockedHostnames []string, errs []error)
	IPs(ctx context.Context,
		blockMalicious, blockAds, blockSurveillance bool,
		additionalBlockedIPs []net.IP, additionalBlockedIPNets []*net.IPNet) (
		blockedIPs []net.IP, blockedIPNets []*net.IPNet, errs []error)
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
