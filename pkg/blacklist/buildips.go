package blacklist

import (
	"bytes"
	"context"
	"net"
	"sort"
)

const (
	adsBlockListIPsURL          = "https://raw.githubusercontent.com/qdm12/files/master/ads-ips.updated"
	maliciousBlockListIPsURL    = "https://raw.githubusercontent.com/qdm12/files/master/malicious-ips.updated"
	surveillanceBlockListIPsURL = "https://raw.githubusercontent.com/qdm12/files/master/surveillance-ips.updated"
)

func (b *builder) IPs(ctx context.Context,
	blockMalicious, blockAds, blockSurveillance bool,
	additionalBlockedIPs []net.IP, additionalBlockedIPNets []*net.IPNet) (
	blockedIPs []net.IP, blockedIPNets []*net.IPNet, errs []error) {
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

	for _, blockedIPNet := range additionalBlockedIPNets {
		uniqueResults[blockedIPNet.String()] = struct{}{}
	}

	blockedIPs = make([]net.IP, 0, len(uniqueResults))
	blockedIPNets = make([]*net.IPNet, 0, len(uniqueResults))

	for s := range uniqueResults {
		ip := net.ParseIP(s)
		if ip != nil {
			blockedIPs = append(blockedIPs, ip)
			continue // TODO with net.ParseCIDR() ?
		}
		_, cidr, err := net.ParseCIDR(s)
		if err == nil && cidr != nil {
			blockedIPNets = append(blockedIPNets, cidr)
			continue
		}
	}

	sort.Slice(blockedIPs, func(i, j int) bool {
		return bytes.Compare([]byte(blockedIPs[i]), []byte(blockedIPs[j])) < 0
	})

	sort.Slice(blockedIPNets, func(i, j int) bool {
		return blockedIPNets[i].String() < blockedIPNets[j].String()
	})

	return blockedIPs, blockedIPNets, errs
}
