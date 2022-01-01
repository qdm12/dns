package log

import (
	"testing"

	"github.com/qdm12/dns/pkg/middlewares/log/format/console"
	"github.com/qdm12/dns/pkg/middlewares/log/logger/noop"
	"github.com/stretchr/testify/assert"
)

func Test_Settings_String(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		settings Settings
		s        string
	}{
		"consoler formatter and noop logger": {
			settings: Settings{
				Formatter: console.New(),
				Logger:    noop.New(),
			},
			s: `Log middleware settings:
├── Logger type: No-op
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
