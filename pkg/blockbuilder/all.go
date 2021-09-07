package blockbuilder

import (
	"context"

	"inet.af/netaddr"
)

func (b *Builder) All(ctx context.Context, settings Settings) (
	result Result) {
	chHostnames := make(chan []string)
	chIPs := make(chan []netaddr.IP)
	chIPPrefixes := make(chan []netaddr.IPPrefix)
	chHostnamesErrors := make(chan []error)
	chIPsErrors := make(chan []error)

	go func() {
		blockedHostnames, errs := b.buildHostnames(ctx,
			settings.BlockMalicious, settings.BlockAds, settings.BlockSurveillance,
			settings.AddBlockedHosts, settings.AllowedHosts)
		chHostnames <- blockedHostnames
		chHostnamesErrors <- errs
	}()

	go func() {
		blockedIPs, blockedIPPrefixes, errs := b.buildIPs(ctx,
			settings.BlockMalicious, settings.BlockAds, settings.BlockSurveillance,
			settings.AllowedIPs, settings.AddBlockedIPs,
			settings.AllowedIPPrefixes, settings.AddBlockedIPPrefixes)
		chIPs <- blockedIPs
		chIPPrefixes <- blockedIPPrefixes
		chIPsErrors <- errs
	}()

	result.BlockedHostnames = <-chHostnames
	result.BlockedIPs = <-chIPs
	result.BlockedIPPrefixes = <-chIPPrefixes

	hostnamesErrors := <-chHostnamesErrors
	result.Errors = append(result.Errors, hostnamesErrors...)
	ipsErrors := <-chIPsErrors
	result.Errors = append(result.Errors, ipsErrors...)

	return result
}
