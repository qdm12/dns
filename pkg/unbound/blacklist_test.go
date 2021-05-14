package unbound

import (
	"testing"

	"github.com/qdm12/dns/pkg/blacklist"
	"github.com/stretchr/testify/assert"
	"inet.af/netaddr"
)

func Test_convertBlockedToConfigLines(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		settings    blacklist.Settings
		configLines []string
	}{
		"none blocked": {
			configLines: []string{},
		},
		"all blocked": {
			settings: blacklist.Settings{
				FqdnHostnames: []string{"sitea", "siteb"},
				IPs: []netaddr.IP{
					netaddr.IPv4(1, 2, 3, 4),
					netaddr.IPv4(4, 3, 2, 1),
				},
				IPPrefixes: []netaddr.IPPrefix{{
					IP:   netaddr.IPv4(5, 5, 5, 5),
					Bits: 16,
				}},
			},
			configLines: []string{
				"  local-zone: \"sitea\" static",
				"  local-zone: \"siteb\" static",
				"  private-address: 1.2.3.4",
				"  private-address: 4.3.2.1",
				"  private-address: 5.5.5.5/16",
			},
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			configLines := convertBlockedToConfigLines(tc.settings)

			assert.Equal(t, tc.configLines, configLines)
		})
	}
}
