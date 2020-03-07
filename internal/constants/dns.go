package constants

import (
	"net"

	"github.com/qdm12/cloudflare-dns-server/internal/models"
)

const (
	// Cloudflare is a DNS over TLS provider
	Cloudflare models.Provider = "cloudflare"
	// Google is a DNS over TLS provider
	Google models.Provider = "google"
	// Quad9 is a DNS over TLS provider
	Quad9 models.Provider = "quad9"
	// Quadrant is a DNS over TLS provider
	Quadrant models.Provider = "quadrant"
	// CleanBrowsing is a DNS over TLS provider
	CleanBrowsing models.Provider = "cleanbrowsing"
	// SecureDNS is a DNS over TLS provider
	SecureDNS models.Provider = "securedns"
	// LibreDNS is a DNS over TLS provider
	LibreDNS models.Provider = "libredns"
	// CIRA is a DNS over TLS provider
	CIRA models.Provider = "cira"
)

// ProviderMapping returns a constant mapping of dns provider name
// to their data such as IP addresses or TLS host name.
func ProviderMapping() map[models.Provider]models.ProviderData {
	return map[models.Provider]models.ProviderData{
		Cloudflare: models.ProviderData{
			IPs:  []net.IP{{1, 1, 1, 1}, {1, 0, 0, 1}},
			Host: models.Host("cloudflare-dns.com"),
		},
		Google: models.ProviderData{
			IPs:  []net.IP{{8, 8, 8, 8}, {8, 8, 4, 4}},
			Host: models.Host("dns.google"),
		},
		Quad9: models.ProviderData{
			IPs:  []net.IP{{9, 9, 9, 9}, {149, 112, 112, 112}},
			Host: models.Host("dns.quad9.net"),
		},
		Quadrant: models.ProviderData{
			IPs:  []net.IP{{12, 159, 2, 159}},
			Host: models.Host("dns-tls.qis.io"),
		},
		CleanBrowsing: models.ProviderData{
			IPs:  []net.IP{{185, 228, 168, 9}, {185, 228, 169, 9}},
			Host: models.Host("security-filter-dns.cleanbrowsing.org"),
		},
		SecureDNS: models.ProviderData{
			IPs:  []net.IP{{146, 185, 167, 43}},
			Host: models.Host("dot.securedns.eu"),
		},
		LibreDNS: models.ProviderData{
			IPs:  []net.IP{{116, 203, 115, 192}},
			Host: models.Host("dot.libredns.gr"),
		},
		CIRA: models.ProviderData{
			IPs:  []net.IP{{149, 112, 121, 20}, {149, 112, 122, 20}},
			Host: models.Host("protected.canadianshield.cira.ca"),
		},
	}
}

// Block lists URLs
const (
	AdsBlockListHostnamesURL          models.URL = "https://raw.githubusercontent.com/qdm12/files/master/ads-hostnames.updated"
	AdsBlockListIPsURL                models.URL = "https://raw.githubusercontent.com/qdm12/files/master/ads-ips.updated"
	MaliciousBlockListHostnamesURL    models.URL = "https://raw.githubusercontent.com/qdm12/files/master/malicious-hostnames.updated"
	MaliciousBlockListIPsURL          models.URL = "https://raw.githubusercontent.com/qdm12/files/master/malicious-ips.updated"
	SurveillanceBlockListHostnamesURL models.URL = "https://raw.githubusercontent.com/qdm12/files/master/surveillance-hostnames.updated"
	SurveillanceBlockListIPsURL       models.URL = "https://raw.githubusercontent.com/qdm12/files/master/surveillance-ips.updated"
)

// DNS certificates to fetch
// TODO obtain from source directly, see qdm12/updated)
const (
	NamedRootURL models.URL = "https://raw.githubusercontent.com/qdm12/files/master/named.root.updated"
	RootKeyURL   models.URL = "https://raw.githubusercontent.com/qdm12/files/master/root.key.updated"
)
