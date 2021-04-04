package unbound

import (
	"net"
	"testing"

	"github.com/qdm12/dns/pkg/blacklist"
	"github.com/stretchr/testify/assert"
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
				IPs:           []net.IP{{1, 2, 3, 4}, {4, 3, 2, 1}},
				IPNets: []*net.IPNet{{
					IP:   net.IP{5, 5, 5, 5},
					Mask: net.IPMask{255, 255, 0, 0},
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
