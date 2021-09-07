package blockbuilder

import (
	"context"

	"inet.af/netaddr"
)

func (b *Builder) All(ctx context.Context, settings Settings) (
	blockedHostnames []string, blockedIPs []netaddr.IP,
	blockedIPPrefixes []netaddr.IPPrefix, errs []error) {
	chHostnames := make(chan []string)
	chIPs := make(chan []netaddr.IP)
	chIPPrefixes := make(chan []netaddr.IPPrefix)
	chErrors := make(chan []error)

	go func() {
		blockedHostnames, errs := b.Hostnames(ctx,
			settings.BlockMalicious, settings.BlockAds, settings.BlockSurveillance,
			settings.AddBlockedHosts, settings.AllowedHosts)
		chHostnames <- blockedHostnames
		chErrors <- errs
	}()

	go func() {
		blockedIPs, blockedIPPrefixes, errs := b.IPs(ctx,
			settings.BlockMalicious, settings.BlockAds, settings.BlockSurveillance,
			settings.AllowedIPs, settings.AddBlockedIPs,
			settings.AllowedIPPrefixes, settings.AddBlockedIPPrefixes)
		chIPs <- blockedIPs
		chIPPrefixes <- blockedIPPrefixes
		chErrors <- errs
	}()

	blockedHostnames = <-chHostnames
	blockedIPs = <-chIPs
	blockedIPPrefixes = <-chIPPrefixes

	routineErrs := <-chErrors
	errs = append(errs, routineErrs...)
	routineErrs = <-chErrors
	errs = append(errs, routineErrs...)

	return blockedHostnames, blockedIPs, blockedIPPrefixes, errs
}
