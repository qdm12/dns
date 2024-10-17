package blockbuilder

import (
	"context"
	"strings"
)

//nolint:lll
const (
	adsBlockListHostnamesURL          = "https://raw.githubusercontent.com/qdm12/files/master/ads-hostnames.updated"
	maliciousBlockListHostnamesURL    = "https://raw.githubusercontent.com/qdm12/files/master/malicious-hostnames.updated"
	surveillanceBlockListHostnamesURL = "https://raw.githubusercontent.com/qdm12/files/master/surveillance-hostnames.updated"
)

func (b *Builder) buildHostnames(ctx context.Context,
	blockMalicious, blockAds, blockSurveillance bool,
	additionalBlockedHostnames, allowedHostnames []string) (
	blockedHostnames []string, errs []error,
) {
	urls := getHostnamesURLs(blockMalicious, blockAds, blockSurveillance)

	uniqueResults, errs := getLists(ctx, b.client, urls)

	for _, blockedHostname := range additionalBlockedHostnames {
		allowed := false
		for _, allowedHostname := range allowedHostnames {
			if blockedHostname == allowedHostname || strings.HasSuffix(blockedHostname, "."+allowedHostname) {
				allowed = true
			}
		}
		if allowed {
			continue
		}
		uniqueResults[blockedHostname] = struct{}{}
	}
	for _, allowedHostname := range allowedHostnames {
		delete(uniqueResults, allowedHostname)
	}

	blockedHostnames = make([]string, 0, len(uniqueResults))
	for result := range uniqueResults {
		blockedHostnames = append(blockedHostnames, result)
	}

	return blockedHostnames, errs
}

func getHostnamesURLs(blockMalicious, blockAds, blockSurveillance bool) (urls []string) {
	const maxURLs = 3
	urls = make([]string, 0, maxURLs)
	if blockMalicious {
		urls = append(urls, string(maliciousBlockListHostnamesURL))
	}
	if blockAds {
		urls = append(urls, string(adsBlockListHostnamesURL))
	}
	if blockSurveillance {
		urls = append(urls, string(surveillanceBlockListHostnamesURL))
	}
	return urls
}
