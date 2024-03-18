package dnssec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_desiredZoneToZoneNames(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		desiredZone string
		zoneNames   []string
	}{
		"root": {
			desiredZone: ".",
			zoneNames:   []string{"."},
		},
		"com": {
			desiredZone: "com.",
			zoneNames:   []string{".", "com."},
		},
		"example.com": {
			desiredZone: "example.com.",
			zoneNames:   []string{".", "com.", "example.com."},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			zoneNames := desiredZoneToZoneNames(testCase.desiredZone)
			assert.Equal(t, testCase.zoneNames, zoneNames)
		})
	}
}
