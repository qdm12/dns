package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Settings_String(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		settings Settings
		s        string
	}{
		"console formatter and noop logger": {
			settings: Settings{
				Format:     "console",
				LoggerType: "noop",
			},
			s: `Log middleware settings:
├── Logger type: Noop
└── Formatter type: Console`,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := testCase.settings.String()

			assert.Equal(t, testCase.s, s)
		})
	}
}
