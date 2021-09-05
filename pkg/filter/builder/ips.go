package builder

import (
	"context"
	"sort"

	"inet.af/netaddr"
)

const (
	adsBlockListIPsURL          = "https://raw.githubusercontent.com/qdm12/files/master/ads-ips.updated"
	maliciousBlockListIPsURL    = "https://raw.githubusercontent.com/qdm12/files/master/malicious-ips.updated"
	surveillanceBlockListIPsURL = "https://raw.githubusercontent.com/qdm12/files/master/surveillance-ips.updated"
)

func (b *builder) IPs(ctx context.Context,
	blockMalicious, blockAds, blockSurveillance bool,
	additionalBlockedIPs []netaddr.IP, additionalBlockedIPPrefixes []netaddr.IPPrefix) (
	blockedIPs []netaddr.IP, blockedIPPrefixes []netaddr.IPPrefix, errs []error) {
	chResults := make(chan []string)
	chError := make(chan error)
	listsLeftToFetch := 0
	if blockMalicious {
		listsLeftToFetch++
		go func() {
			results, err := getList(ctx, b.client, string(maliciousBlockListIPsURL))
			chResults <- results
			chError <- err
		}()
	}
	if blockAds {
		listsLeftToFetch++
		go func() {
			results, err := getList(ctx, b.client, string(adsBlockListIPsURL))
			chResults <- results
			chError <- err
		}()
	}
	if blockSurveillance {
		listsLeftToFetch++
		go func() {
			results, err := getList(ctx, b.client, string(surveillanceBlockListIPsURL))
			chResults <- results
			chError <- err
		}()
	}
	uniqueResults := make(map[string]struct{})
	for listsLeftToFetch > 0 {
		select {
		case results := <-chResults:
			for _, result := range results {
				uniqueResults[result] = struct{}{}
			}
		case err := <-chError:
			listsLeftToFetch--
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	for _, blockedIP := range additionalBlockedIPs {
		uniqueResults[blockedIP.String()] = struct{}{}
	}

	for _, blockedIPPrefix := range additionalBlockedIPPrefixes {
		uniqueResults[blockedIPPrefix.String()] = struct{}{}
	}

	blockedIPs = make([]netaddr.IP, 0, len(uniqueResults))
	blockedIPPrefixes = make([]netaddr.IPPrefix, 0, len(uniqueResults))

	for s := range uniqueResults {
		ip, err := netaddr.ParseIP(s)
		if err == nil {
			blockedIPs = append(blockedIPs, ip)
			continue
		}

		ipPrefix, err := netaddr.ParseIPPrefix(s)
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

	return blockedIPs, blockedIPPrefixes, errs
}
