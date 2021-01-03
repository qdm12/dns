package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Settings_Lines(t *testing.T) {
	t.Parallel()
	testCases := map[string]struct {
		settings Settings
		lines    []string
	}{
		"empty settings": {
			lines: []string{
				"DNS over TLS provider:",
				"Listening port: 0",
				"Caching: disabled",
				"IPv4 resolution: disabled",
				"IPv6 resolution: disabled",
				"Verbosity level: 0/5",
				"Verbosity details level: 0/4",
				"Validation log level: 0/2",
				"Blocked hostnames:",
				"Blocked IP addresses:",
				"Allowed hostnames:",
			},
		},
		"full settings": {
			settings: Settings{
				Providers:             []string{"quad9", "cloudflare"},
				ListeningPort:         53,
				Caching:               true,
				IPv4:                  true,
				IPv6:                  true,
				VerbosityLevel:        1,
				VerbosityDetailsLevel: 2,
				ValidationLogLevel:    3,
				BlockedHostnames:      []string{"hostname 1", "hostname 2"},
				BlockedIPs:            []string{"1.1.1.2", "2.2.2.2"},
				AllowedHostnames:      []string{"hostname 3", "hostname 4"},
			},
			lines: []string{
				"DNS over TLS provider:",
				" |--quad9",
				" |--cloudflare",
				"Listening port: 53",
				"Caching: enabled",
				"IPv4 resolution: enabled",
				"IPv6 resolution: enabled",
				"Verbosity level: 1/5",
				"Verbosity details level: 2/4",
				"Validation log level: 3/2",
				"Blocked hostnames:",
				" |--hostname 1",
				" |--hostname 2",
				"Blocked IP addresses:",
				" |--1.1.1.2",
				" |--2.2.2.2",
				"Allowed hostnames:",
				" |--hostname 3",
				" |--hostname 4",
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
