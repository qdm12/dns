package env

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setTestEnv is used to set environment variables in
// parallel tests.
func setTestEnv(t *testing.T, key, value string) {
	t.Helper()
	existing := os.Getenv(key)
	err := os.Setenv(key, value) //nolint:tenv
	t.Cleanup(func() {
		err = os.Setenv(key, existing)
		assert.NoError(t, err)
	})
	require.NoError(t, err)
}

func Test_envToDurationPtr(t *testing.T) {
	t.Parallel()

	durationPtr := func(d time.Duration) *time.Duration { return &d }

	testCases := map[string]struct {
		envKey       string
		envValue     string
		d            *time.Duration
		errorMessage string
	}{
		"empty": {
			envKey: "DURATION_EMPTY",
		},
		"zero": {
			envKey:   "DURATION_ZERO",
			envValue: "0",
			d:        durationPtr(0),
		},
		"one second": {
			envKey:   "DURATION_ONE_SECOND",
			envValue: "1s",
			d:        durationPtr(time.Second),
		},
		"parse error": {
			envKey:       "DURATION_MALFORMED",
			envValue:     "x",
			errorMessage: "time: invalid duration \"x\"",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			setTestEnv(t, testCase.envKey, testCase.envValue)

			d, err := envToDurationPtr(testCase.envKey)

			assert.Equal(t, testCase.d, d)
			if testCase.errorMessage != "" {
				assert.EqualError(t, err, testCase.errorMessage)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
