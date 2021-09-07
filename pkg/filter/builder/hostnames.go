package builder

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

func (b *Builder) Hostnames(ctx context.Context,
	blockMalicious, blockAds, blockSurveillance bool,
	additionalBlockedHostnames, allowedHostnames []string) (
	blockedHostnames []string, errs []error) {
	chResults := make(chan []string)
	chError := make(chan error)
	listsLeftToFetch := 0
	if blockMalicious {
		listsLeftToFetch++
		go func() {
			results, err := getList(ctx, b.client, string(maliciousBlockListHostnamesURL))
			chResults <- results
			chError <- err
		}()
	}
	if blockAds {
		listsLeftToFetch++
		go func() {
			results, err := getList(ctx, b.client, string(adsBlockListHostnamesURL))
			chResults <- results
			chError <- err
		}()
	}
	if blockSurveillance {
		listsLeftToFetch++
		go func() {
			results, err := getList(ctx, b.client, string(surveillanceBlockListHostnamesURL))
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
