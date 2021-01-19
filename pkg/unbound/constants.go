package unbound

import (
	"net"

	"github.com/qdm12/dns/internal/models"
)

const (
	unboundConfigFilename = "unbound.conf"
	rootHints             = "root.hints"
	rootKey               = "root.key"
)

const (
	resolvConfFilepath = "/etc/resolv.conf"
)

const (
	// Cloudflare is a DNS over TLS provider.
	Cloudflare = "cloudflare"
	// CloudflareSecurity is a DNS over TLS provider blocking malware.
	CloudflareSecurity = "cloudflare.security"
	// CloudflareFamily is a DNS over TLS provider blocking malware and adult content.
	CloudflareFamily = "cloudflare.family"
	// Google is a DNS over TLS provider.
	Google = "google"
	// Quad9 is a DNS over TLS provider.
	Quad9 = "quad9"
	// Quadrant is a DNS over TLS provider.
	Quadrant = "quadrant"
	// CleanBrowsing is a DNS over TLS provider.
	CleanBrowsing = "cleanbrowsing"
	// CleanBrowsingFamily is a DNS over TLS provider blocking malware, adult content and mixed content.
	CleanBrowsingFamily = "cleanbrowsing.family"
	// CleanBrowsingAdult is a DNS over TLS provider blocking adult content.
	CleanBrowsingAdult = "cleanbrowsing.adult"
	// LibreDNS is a DNS over TLS provider.
	LibreDNS = "libredns"
	// CIRA is a DNS over TLS provider.
	CIRA = "cira"
)

func GetProviderData(provider string) (data models.ProviderData, ok bool) {
	mapping := map[string]models.ProviderData{
		Cloudflare: {
			IPs: []net.IP{
				{1, 1, 1, 1},
				{1, 0, 0, 1},
				{0x26, 0x6, 0x47, 0x0, 0x47, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x11, 0x11},
				{0x26, 0x6, 0x47, 0x0, 0x47, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x10, 0x01},
			},
			SupportsTLS:    true,
			SupportsDNSSEC: true,
			SupportsIPv6:   true,
			Host:           models.Host("cloudflare-dns.com"),
		},
		CloudflareSecurity: {
			IPs: []net.IP{
				{1, 1, 1, 2},
				{1, 0, 0, 2},
				{0x26, 0x6, 0x47, 0x0, 0x47, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x11, 0x12},
				{0x26, 0x6, 0x47, 0x0, 0x47, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x10, 0x02},
			},
			SupportsTLS:    true,
			SupportsDNSSEC: true,
			SupportsIPv6:   true,
			Host:           models.Host("security.cloudflare-dns.com"),
		},
		CloudflareFamily: {
			IPs: []net.IP{
				{1, 1, 1, 3},
				{1, 0, 0, 3},
				{0x26, 0x6, 0x47, 0x0, 0x47, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x11, 0x13},
				{0x26, 0x6, 0x47, 0x0, 0x47, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x10, 0x03},
			},
			SupportsTLS:    true,
			SupportsDNSSEC: true,
			SupportsIPv6:   true,
			Host:           models.Host("family.cloudflare-dns.com"),
		},
		Google: {
			IPs: []net.IP{
				{8, 8, 8, 8},
				{8, 8, 4, 4},
				{0x20, 0x1, 0x48, 0x60, 0x48, 0x60, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x88, 0x88},
				{0x20, 0x1, 0x48, 0x60, 0x48, 0x60, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x88, 0x44},
			},
			SupportsTLS:    true,
			SupportsDNSSEC: true,
			SupportsIPv6:   true,
			Host:           models.Host("dns.google"),
		},
		Quad9: {
			IPs: []net.IP{
				{9, 9, 9, 9},
				{149, 112, 112, 112},
				{0x26, 0x20, 0x0, 0xfe, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe},
				{0x26, 0x20, 0x0, 0xfe, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x9},
			},
			SupportsTLS:    true,
			SupportsDNSSEC: true,
			SupportsIPv6:   true,
			Host:           models.Host("dns.quad9.net"),
		},
		Quadrant: {
			IPs: []net.IP{
				{12, 159, 2, 159},
				{0x20, 0x1, 0x18, 0x90, 0x14, 0xc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x59},
			},
			SupportsTLS:    true,
			SupportsDNSSEC: true,
			SupportsIPv6:   true,
			Host:           models.Host("dns-tls.qis.io"),
		},
		CleanBrowsing: {
			IPs: []net.IP{
				{185, 228, 168, 9},
				{185, 228, 169, 9},
				{0x2a, 0xd, 0x2a, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2},
				{0x2a, 0xd, 0x2a, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2},
			},
			SupportsTLS:    true,
			SupportsDNSSEC: true,
			SupportsIPv6:   true,
			Host:           models.Host("security-filter-dns.cleanbrowsing.org"),
		},
		CleanBrowsingFamily: {
			IPs: []net.IP{
				{185, 228, 168, 168},
				{185, 228, 169, 168},
				{0x2a, 0xd, 0x2a, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				{0x2a, 0xd, 0x2a, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			},
			SupportsTLS:    true,
			SupportsDNSSEC: true,
			SupportsIPv6:   true,
			Host:           models.Host("family-filter-dns.cleanbrowsing.org"),
		},
		CleanBrowsingAdult: {
			IPs: []net.IP{
				{185, 228, 168, 10},
				{185, 228, 169, 11},
				{0x2a, 0xd, 0x2a, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
				{0x2a, 0xd, 0x2a, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
			},
			SupportsTLS:    true,
			SupportsDNSSEC: true,
			SupportsIPv6:   true,
			Host:           models.Host("adult-filter-dns.cleanbrowsing.org"),
		},
		LibreDNS: {
			IPs:         []net.IP{{116, 202, 176, 26}},
			Host:        models.Host("dot.libredns.gr"),
			SupportsTLS: true,
		},
		CIRA: {
			IPs: []net.IP{
				{149, 112, 121, 20},
				{149, 112, 122, 20},
				{0x26, 0x20, 0x1, 0xa, 0x80, 0xbb, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20},
				{0x26, 0x20, 0x1, 0xa, 0x80, 0xbc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20},
			},
			SupportsTLS:    true,
			SupportsDNSSEC: true,
			Host:           models.Host("protected.canadianshield.cira.ca"),
		},
	}
	data, ok = mapping[provider]
	return data, ok
}

//nolint:lll
const (
	adsBlockListHostnamesURL          models.URL = "https://raw.githubusercontent.com/qdm12/files/master/ads-hostnames.updated"
	adsBlockListIPsURL                models.URL = "https://raw.githubusercontent.com/qdm12/files/master/ads-ips.updated"
	maliciousBlockListHostnamesURL    models.URL = "https://raw.githubusercontent.com/qdm12/files/master/malicious-hostnames.updated"
	maliciousBlockListIPsURL          models.URL = "https://raw.githubusercontent.com/qdm12/files/master/malicious-ips.updated"
	surveillanceBlockListHostnamesURL models.URL = "https://raw.githubusercontent.com/qdm12/files/master/surveillance-hostnames.updated"
	surveillanceBlockListIPsURL       models.URL = "https://raw.githubusercontent.com/qdm12/files/master/surveillance-ips.updated"
)
