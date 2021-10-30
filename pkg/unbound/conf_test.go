package unbound

import (
	"strings"
	"testing"

	"github.com/qdm12/dns/pkg/provider"
	"github.com/stretchr/testify/assert"
	"inet.af/netaddr"
)

func Test_generateUnboundConf(t *testing.T) {
	t.Parallel()
	settings := Settings{
		Providers: []provider.Provider{
			provider.Cloudflare(),
			provider.Quad9(),
		},
		VerbosityLevel:     2,
		ValidationLogLevel: 3,
		ListeningPort:      53,
		IPv4:               true,
		IPv6:               true,
		AccessControl: AccessControlSettings{
			Allowed: []netaddr.IPPrefix{{IP: netaddr.IPv4(0, 0, 0, 0)}},
		},
	}
	lines := generateUnboundConf(settings,
		[]string{
			`  local-zone: "b" static`,
			`  local-zone: "c" static`,
			"  private-address: c",
			"  private-address: d",
		},
		"/unbound",
		"/unbound/ca-certificates.crt",
		"user",
	)
	expected := `
server:
  access-control: 0.0.0.0/0 allow
  cache-max-ttl: 9000
  cache-min-ttl: 0
  do-ip4: yes
  do-ip6: yes
  harden-algo-downgrade: yes
  harden-below-nxdomain: yes
  harden-referral-path: yes
  hide-identity: yes
  hide-version: yes
  include: "/unbound/include.conf"
  interface: 0.0.0.0
  key-cache-size: 32m
  key-cache-slabs: 4
  msg-cache-size: 8m
  msg-cache-slabs: 4
  num-threads: 2
  port: 53
  prefetch-key: yes
  prefetch: yes
  root-hints: "/unbound/root.hints"
  rrset-cache-size: 8m
  rrset-cache-slabs: 4
  rrset-roundrobin: yes
  tls-cert-bundle: "/unbound/ca-certificates.crt"
  trust-anchor-file: "/unbound/root.key"
  use-syslog: no
  username: "user"
  val-log-level: 3
  verbosity: 2
  local-zone: "b" static
  local-zone: "c" static
  private-address: c
  private-address: d
forward-zone:
  forward-no-cache: yes
  forward-tls-upstream: yes
  name: "."
  forward-addr: 1.1.1.1@853#cloudflare-dns.com
  forward-addr: 1.0.0.1@853#cloudflare-dns.com
  forward-addr: 2606:4700:4700::1111@853#cloudflare-dns.com
  forward-addr: 2606:4700:4700::1001@853#cloudflare-dns.com
  forward-addr: 9.9.9.9@853#dns.quad9.net
  forward-addr: 149.112.112.112@853#dns.quad9.net
  forward-addr: 2620:fe::fe@853#dns.quad9.net
  forward-addr: 2620:fe::9@853#dns.quad9.net`
	assert.Equal(t, expected, "\n"+strings.Join(lines, "\n"))
}
