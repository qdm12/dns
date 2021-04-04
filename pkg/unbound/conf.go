package unbound

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/qdm12/golibs/os"
)

func (c *configurator) MakeUnboundConf(settings Settings) (err error) {
	configFilepath := filepath.Join(c.unboundEtcDir, unboundConfigFilename)
	file, err := c.openFile(configFilepath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	blacklistLines := convertBlockedToConfigLines(settings.Blacklist)

	lines := generateUnboundConf(settings, blacklistLines,
		c.unboundEtcDir, c.cacertsPath, settings.Username)
	_, err = file.WriteString(strings.Join(lines, "\n"))
	if err != nil {
		_ = file.Close()
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return nil
}

// generateUnboundConf generates an Unbound configuration from the user provided settings.
func generateUnboundConf(settings Settings, blacklistLines []string,
	unboundDir, cacertsPath, username string) (
	lines []string) {
	const (
		yes = "yes"
		no  = "no"
	)
	ipv4, ipv6 := no, no
	if settings.IPv4 {
		ipv4 = yes
	}
	if settings.IPv6 {
		ipv6 = yes
	}
	serverLines := []string{
		// Logging
		"verbosity: " + strconv.Itoa(int(settings.VerbosityLevel)),
		"val-log-level: " + strconv.Itoa(int(settings.ValidationLogLevel)),
		"use-syslog: no",
		// Performance
		"num-threads: 2",
		"prefetch: yes",
		"prefetch-key: yes",
		"key-cache-size: 32m",
		"key-cache-slabs: 4",
		"msg-cache-size: 8m",
		"msg-cache-slabs: 4",
		"rrset-cache-size: 8m",
		"rrset-cache-slabs: 4",
		"cache-min-ttl: 3600",
		"cache-max-ttl: 9000",
		// Privacy
		"rrset-roundrobin: yes",
		"hide-identity: yes",
		"hide-version: yes",
		// Security
		`tls-cert-bundle: "` + cacertsPath + `"`,
		`root-hints: "` + filepath.Join(unboundDir, rootHints) + `"`,
		"harden-below-nxdomain: yes",
		"harden-referral-path: yes",
		"harden-algo-downgrade: yes",
		// Network
		"do-ip4: " + ipv4,
		"do-ip6: " + ipv6,
		"interface: 0.0.0.0",
		"port: " + strconv.Itoa(int(settings.ListeningPort)),
		// Other
		`username: "` + username + `"`,
		`trust-anchor-file: "` + filepath.Join(unboundDir, rootKey) + `"`, // DNSSEC trust anchor file
		`include: "` + filepath.Join(unboundDir, includeConfFilename) + `"`,
	}

	// Access control
	for _, subnet := range settings.AccessControl.Allowed {
		line := "access-control: " + subnet.String() + " allow"
		serverLines = append(serverLines, line)
	}

	serverLines = ensureIndentLines(serverLines)
	sort.Slice(serverLines, func(i, j int) bool {
		return serverLines[i] < serverLines[j]
	})

	blacklistLines = ensureIndentLines(blacklistLines)

	lines = append(lines, "server:")
	lines = append(lines, serverLines...)
	lines = append(lines, blacklistLines...)

	// Forward zone
	lines = append(lines, "forward-zone:")
	forwardZoneLines := []string{
		`name: "."`,
		"forward-tls-upstream: yes",
	}
	cachingLine := "forward-no-cache: yes"
	if settings.Caching {
		cachingLine = "forward-no-cache: no"
	}
	forwardZoneLines = append(forwardZoneLines, cachingLine)

	sort.Slice(forwardZoneLines, func(i, j int) bool {
		return forwardZoneLines[i] < forwardZoneLines[j]
	})

	for _, provider := range settings.Providers {
		dotServer := provider.DoT()
		ips := append(dotServer.IPv4, dotServer.IPv6...)
		for _, IP := range ips {
			forwardZoneLines = append(forwardZoneLines,
				fmt.Sprintf("forward-addr: %s@853#%s", IP.String(), dotServer.Name))
		}
	}

	forwardZoneLines = ensureIndentLines(forwardZoneLines)

	lines = append(lines, forwardZoneLines...)
	return lines
}

func ensureIndentLines(lines []string) []string {
	const spaces = 2
	indent := strings.Repeat(" ", spaces)
	for i := range lines {
		if !strings.HasPrefix(lines[i], indent) {
			lines[i] = indent + lines[i]
		}
	}
	return lines
}
