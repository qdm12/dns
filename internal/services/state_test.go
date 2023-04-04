package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_State_String(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		state  State
		result string
	}{
		"stopped": {
			state:  StateStopped,
			result: "stopped",
		},
		"starting": {
			state:  StateStarting,
			result: "starting",
		},
		"running": {
			state:  StateRunning,
			result: "running",
		},
		"stopping": {
			state:  StateStopping,
			result: "stopping",
		},
		"crashed": {
			state:  StateCrashed,
			result: "crashed",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := testCase.state.String()

			assert.Equal(t, testCase.result, result)
		})
	}
}
