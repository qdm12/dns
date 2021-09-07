package config

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/qdm12/golibs/logging"
	"github.com/qdm12/golibs/params"
	"github.com/qdm12/golibs/params/mock_params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getLogSettings(t *testing.T) {
	t.Parallel()

	errDummy := errors.New("dummy")

	type levelResult struct {
		level logging.Level
		err   error
	}

	type onOffResult struct {
		on  bool
		err error
	}

	testCases := map[string]struct {
		logLevel             levelResult
		logRequests          onOffResult
		logResponses         onOffResult
		logRequestsResponses onOffResult
		settings             Log
		err                  error
	}{
		"defaults": {},
		"log level error": {
			logLevel: levelResult{err: errDummy},
			err:      errors.New("environment variable LOG_LEVEL: dummy"),
		},
		"log requests error": {
			logRequests: onOffResult{err: errDummy},
			err:         errors.New("environment variable LOG_REQUESTS: dummy"),
		},
		"log responses error": {
			logResponses: onOffResult{err: errDummy},
			err:          errors.New("environment variable LOG_RESPONSES: dummy"),
		},
		"log requests responses error": {
			logRequestsResponses: onOffResult{err: errDummy},
			err:                  errors.New("environment variable LOG_REQUESTS_RESPONSES: dummy"),
		},
		"all set": {
			logLevel:             levelResult{level: logging.LevelInfo},
			logRequests:          onOffResult{on: true},
			logResponses:         onOffResult{on: true},
			logRequestsResponses: onOffResult{on: true},
			settings: Log{
				Level:                logging.LevelInfo,
				LogRequests:          true,
				LogResponses:         true,
				LogRequestsResponses: true,
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			assignableDefault := gomock.AssignableToTypeOf(params.Default(""))

			env := mock_params.NewMockInterface(ctrl)
			env.EXPECT().
				LogLevel("LOG_LEVEL", assignableDefault).
				Return(testCase.logLevel.level, testCase.logLevel.err)

			if testCase.logLevel.err == nil {
				env.EXPECT().
					OnOff("LOG_REQUESTS", assignableDefault).
					Return(testCase.logRequests.on, testCase.logRequests.err)
			}

			if testCase.logLevel.err == nil &&
				testCase.logRequests.err == nil {
				env.EXPECT().
					OnOff("LOG_RESPONSES", assignableDefault).
					Return(testCase.logResponses.on, testCase.logResponses.err)
			}

			if testCase.logLevel.err == nil &&
				testCase.logRequests.err == nil &&
				testCase.logResponses.err == nil {
				env.EXPECT().
					OnOff("LOG_REQUESTS_RESPONSES", assignableDefault).
					Return(testCase.logRequestsResponses.on, testCase.logRequestsResponses.err)
			}

			settings, err := getLogSettings(env)

			if testCase.err != nil {
				require.Error(t, err)
				assert.Equal(t, testCase.err.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.settings, settings)
		})
	}
}
