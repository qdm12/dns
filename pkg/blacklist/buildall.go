package blacklist

import (
	"context"
	"net"
)

func (b *builder) All(ctx context.Context,
	blockMalicious, blockAds, blockSurveillance bool,
	additionalBlockedHostnames, allowedHostnames []string,
	additionalBlockedIPs []net.IP, additionalBlockedIPNets []*net.IPNet) (
	blockedHostnames []string, blockedIPs []net.IP,
	blockedIPNets []*net.IPNet, errs []error) {
	chHostnames := make(chan []string)
	chIPs := make(chan []net.IP)
	chIPNets := make(chan []*net.IPNet)
	chErrors := make(chan []error)

	go func() {
		blockedHostnames, errs := b.Hostnames(ctx,
			blockMalicious, blockAds, blockSurveillance,
			additionalBlockedHostnames, allowedHostnames)
		chHostnames <- blockedHostnames
		chErrors <- errs
	}()

	go func() {
		blockedIPs, blockedIPNets, errs := b.IPs(ctx,
			blockMalicious, blockAds, blockSurveillance,
			additionalBlockedIPs, additionalBlockedIPNets)
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
