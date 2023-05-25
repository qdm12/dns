package blockbuilder

import (
	"context"
	"net/netip"
)

func (b *Builder) BuildAll(ctx context.Context) (result Result) {
	chHostnames := make(chan []string)
	chIPs := make(chan []netip.Addr)
	chIPPrefixes := make(chan []netip.Prefix)
	chHostnamesErrors := make(chan []error)
	chIPsErrors := make(chan []error)

	go func() {
		blockedHostnames, errs := b.buildHostnames(ctx,
			b.blockMalicious, b.blockAds, b.blockSurveillance,
			b.addBlockedHosts, b.allowedHosts)
		chHostnames <- blockedHostnames
		chHostnamesErrors <- errs
	}()

	go func() {
		blockedIPs, blockedIPPrefixes, errs := b.buildIPs(ctx,
			b.blockMalicious, b.blockAds, b.blockSurveillance,
			b.allowedIPs, b.addBlockedIPs,
			b.allowedIPPrefixes, b.addBlockedIPPrefixes)
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
