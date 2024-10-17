package substituter

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Settings_SetDefaults(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		initialSettings  Settings
		expectedSettings Settings
	}{
		"empty_settings": {},
		"settings_need_defaults": {
			initialSettings: Settings{
				Substitutions: []Substitution{
					{Name: "google.com", IPs: []netip.Addr{netip.MustParseAddr("1.2.3.4")}},
				},
			},
			expectedSettings: Settings{
				Substitutions: []Substitution{
					{Name: "google.com", Type: "A", Class: "IN", IPs: []netip.Addr{netip.MustParseAddr("1.2.3.4")}, TTL: 300},
				},
			},
		},
		"settings_filled": {
			initialSettings: Settings{
				Substitutions: []Substitution{
					{Name: "google.com", Type: "A", Class: "IN", IPs: []netip.Addr{netip.MustParseAddr("1.2.3.4")}, TTL: 300},
				},
			},
			expectedSettings: Settings{
				Substitutions: []Substitution{
					{Name: "google.com", Type: "A", Class: "IN", IPs: []netip.Addr{netip.MustParseAddr("1.2.3.4")}, TTL: 300},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			settings := testCase.initialSettings
			settings.SetDefaults()

			assert.Equal(t, testCase.expectedSettings, settings)
		})
	}
}

func Test_Settings_Validate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		settings   Settings
		errWrapped error
		errMessage string
	}{
		"no_error": {
			settings: Settings{},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := testCase.settings.Validate()

			assert.ErrorIs(t, err, testCase.errWrapped)
			if testCase.errWrapped != nil {
				assert.EqualError(t, err, testCase.errMessage)
			}
		})
	}
}

func Test_Settings_String(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		settings Settings
		s        string
	}{
		"empty_settings": {
			settings: Settings{},
			s:        `Substitute middleware: disabled`,
		},
		"one_substitution": {
			settings: Settings{
				Substitutions: []Substitution{
					{Name: "google.com", Type: "A", Class: "IN", IPs: []netip.Addr{netip.MustParseAddr("1.2.3.4")}, TTL: 300},
				},
			},
			s: `Substitute middleware settings:
└── Substitutions:
    └── google.com A IN -> 1.2.3.4 with ttl 300`,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := testCase.settings.String()

			assert.Equal(t, testCase.s, s)
		})
	}
}
