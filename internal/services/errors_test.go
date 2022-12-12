package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_serviceError(t *testing.T) {
	t.Parallel()
	errTest := errors.New("test error")

	testCases := map[string]struct {
		serviceError serviceError
		errString    string
		errUnwrapped error
		panicValue   string
	}{
		"no err set panics": {
			serviceError: serviceError{
				format:      errorFormatCrash,
				serviceName: "A",
			},
			panicValue: "cannot have nil error in serviceError",
		},
		"error set": {
			serviceError: serviceError{
				format:      errorFormatCrash,
				serviceName: "A",
				err:         errTest,
			},
			errString:    "A crashed: test error",
			errUnwrapped: errTest,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if testCase.panicValue != "" {
				assert.PanicsWithValue(t, testCase.panicValue, func() {
					_ = testCase.serviceError.Error()
				})
				return
			}

			assert.ErrorIs(t, testCase.serviceError, testCase.errUnwrapped)
			assert.EqualError(t, testCase.serviceError, testCase.errString)
		})
	}
}
