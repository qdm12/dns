package log

import (
	"testing"

	formatnoop "github.com/qdm12/dns/v2/pkg/middlewares/log/format/noop"
	lognoop "github.com/qdm12/dns/v2/pkg/middlewares/log/logger/noop"
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
				Formatter: formatnoop.New(),
				Logger:    lognoop.New(),
			},
			s: `Log middleware settings:
├── Logger type: noop.Logger
└── Formatter type: noop.Formatter`,
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
