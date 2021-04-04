package blacklist

import (
	"context"
	"net"
)

func (b *builder) All(ctx context.Context, settings BuilderSettings) (
	blockedHostnames []string, blockedIPs []net.IP,
	blockedIPNets []*net.IPNet, errs []error) {
	chHostnames := make(chan []string)
	chIPs := make(chan []net.IP)
	chIPNets := make(chan []*net.IPNet)
	chErrors := make(chan []error)

	go func() {
		blockedHostnames, errs := b.Hostnames(ctx,
			settings.BlockMalicious, settings.BlockAds, settings.BlockSurveillance,
			settings.AddBlockedHosts, settings.AllowedHosts)
		chHostnames <- blockedHostnames
		chErrors <- errs
	}()

	go func() {
		blockedIPs, blockedIPNets, errs := b.IPs(ctx,
			settings.BlockMalicious, settings.BlockAds, settings.BlockSurveillance,
			settings.AddBlockedIPs, settings.AddBlockedIPNets)
		chIPs <- blockedIPs
		chIPNets <- blockedIPNets
		chErrors <- errs
	}()

	blockedHostnames = <-chHostnames
	blockedIPs = <-chIPs
	blockedIPNets = <-chIPNets

	routineErrs := <-chErrors
	errs = append(errs, routineErrs...)
	routineErrs = <-chErrors
	errs = append(errs, routineErrs...)

	return blockedHostnames, blockedIPs, blockedIPNets, errs
}
