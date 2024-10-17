package update

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Settings_BlockHostnames(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		initialSettings Settings
		hostnames       []string
		finalSettings   Settings
	}{
		"nothing": {
			finalSettings: Settings{
				FqdnHostnames: []string{},
			},
		},
		"insert new first ones": {
			hostnames: []string{"abc.com", "def.co.uk"},
			finalSettings: Settings{
				FqdnHostnames: []string{"abc.com.", "def.co.uk."},
			},
		},
		"override": {
			initialSettings: Settings{
				FqdnHostnames: []string{"01.com.", "abc.com."},
			},
			hostnames: []string{"abc.com", "def.co.uk"},
			finalSettings: Settings{
				FqdnHostnames: []string{"abc.com.", "def.co.uk."},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			settings := testCase.initialSettings

			settings.BlockHostnames(testCase.hostnames)

			assert.Equal(t, testCase.finalSettings, settings)
		})
	}
}
