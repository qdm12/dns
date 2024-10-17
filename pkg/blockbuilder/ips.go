package blockbuilder

import (
	"context"
	"net/netip"
	"sort"
)

const (
	adsBlockListIPsURL          = "https://raw.githubusercontent.com/qdm12/files/master/ads-ips.updated"
	maliciousBlockListIPsURL    = "https://raw.githubusercontent.com/qdm12/files/master/malicious-ips.updated"
	surveillanceBlockListIPsURL = "https://raw.githubusercontent.com/qdm12/files/master/surveillance-ips.updated"
)

func (b *Builder) buildIPs(ctx context.Context,
	blockMalicious, blockAds, blockSurveillance bool,
	allowedIPs, additionalBlockedIPs []netip.Addr,
	allowedIPPrefixes, additionalBlockedIPPrefixes []netip.Prefix) (
	blockedIPs []netip.Addr, blockedIPPrefixes []netip.Prefix, errs []error,
) {
	urls := getIPsURLs(blockMalicious, blockAds, blockSurveillance)

	uniqueResults, errs := getLists(ctx, b.client, urls)

	for _, blockedIP := range additionalBlockedIPs {
		uniqueResults[blockedIP.String()] = struct{}{}
	}
	for _, allowedIP := range allowedIPs {
		delete(uniqueResults, allowedIP.String())
	}

	for _, blockedIPPrefix := range additionalBlockedIPPrefixes {
		uniqueResults[blockedIPPrefix.String()] = struct{}{}
	}
	for _, allowedIPPrefix := range allowedIPPrefixes {
		delete(uniqueResults, allowedIPPrefix.String())
	}

	blockedIPs, blockedIPPrefixes = parseIPStrings(uniqueResults)

	return blockedIPs, blockedIPPrefixes, errs
}

func getIPsURLs(blockMalicious, blockAds, blockSurveillance bool) (urls []string) {
	const maxURLs = 3
	urls = make([]string, 0, maxURLs)
	if blockMalicious {
		urls = append(urls, string(maliciousBlockListIPsURL))
	}
	if blockAds {
		urls = append(urls, string(adsBlockListIPsURL))
	}
	if blockSurveillance {
		urls = append(urls, string(surveillanceBlockListIPsURL))
	}
	return urls
}

func parseIPStrings(uniqueResults map[string]struct{}) (
	blockedIPs []netip.Addr, blockedIPPrefixes []netip.Prefix,
) {
	blockedIPs = make([]netip.Addr, 0, len(uniqueResults))
	blockedIPPrefixes = make([]netip.Prefix, 0, len(uniqueResults))

	for s := range uniqueResults {
		ip, err := netip.ParseAddr(s)
		if err == nil {
			blockedIPs = append(blockedIPs, ip)
			continue
		}

		ipPrefix, err := netip.ParsePrefix(s)
		if err == nil {
			blockedIPPrefixes = append(blockedIPPrefixes, ipPrefix)
			continue
		}
	}

	sort.Slice(blockedIPs, func(i, j int) bool {
		return blockedIPs[i].Compare(blockedIPs[j]) < 0
	})

	sort.Slice(blockedIPPrefixes, func(i, j int) bool {
		return blockedIPPrefixes[i].String() < blockedIPPrefixes[j].String()
	})

	return blockedIPs, blockedIPPrefixes
}
