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
		"empty settings": {
			s: " |--Status: disabled",
		},
		"log requests only": {
			settings: Settings{
				LogRequests: true,
			},
			s: " |--Log requests: on",
		},
		"log responses only": {
			settings: Settings{
				LogResponses: true,
			},
			s: " |--Log responses: on",
		},
		"log requests and responses": {
			settings: Settings{
				LogRequests:  true,
				LogResponses: true,
			},
			s: ` |--Log requests: on
 |--Log responses: on`,
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
