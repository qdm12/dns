package unbound

import (
	"testing"

	"github.com/qdm12/dns/pkg/provider"
	"github.com/stretchr/testify/assert"
	"inet.af/netaddr"
)

func Test_Settings_Lines(t *testing.T) {
	t.Parallel()
	testCases := map[string]struct {
		settings Settings
		lines    []string
	}{
		"empty settings": {
			lines: []string{
				" |--DNS over TLS providers:",
				" |--Listening port: 0",
				" |--Access control:",
				"     |--Allowed:",
				" |--Caching: disabled",
				" |--IPv4 resolution: disabled",
				" |--IPv6 resolution: disabled",
				" |--Verbosity level: 0/5",
				" |--Verbosity details level: 0/4",
				" |--Validation log level: 0/2",
				" |--Username: ",
			},
		},
		"full settings": {
			settings: Settings{
				Providers: []provider.Provider{
					provider.Quad9(),
					provider.Cloudflare(),
				},
				ListeningPort:         53,
				Caching:               true,
				IPv4:                  true,
				IPv6:                  true,
				VerbosityLevel:        1,
				VerbosityDetailsLevel: 2,
				ValidationLogLevel:    3,
				AccessControl: AccessControlSettings{
					Allowed: []netaddr.IPPrefix{{IP: netaddr.IPv4(0, 0, 0, 0)}},
				},
				Username: "username",
			},
			lines: []string{
				" |--DNS over TLS providers:",
				"     |--Quad9",
				"     |--Cloudflare",
				" |--Listening port: 53",
				" |--Access control:",
				"     |--Allowed:",
				"         |--0.0.0.0/0",
				" |--Caching: enabled",
				" |--IPv4 resolution: enabled",
				" |--IPv6 resolution: enabled",
				" |--Verbosity level: 1/5",
				" |--Verbosity details level: 2/4",
				" |--Validation log level: 3/2",
				" |--Username: username",
			},
		},
	}
	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			lines := testCase.settings.Lines()
			assert.Equal(t, testCase.lines, lines)
		})
	}
}
