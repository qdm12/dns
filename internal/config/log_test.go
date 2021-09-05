package config

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/qdm12/dns/pkg/middlewares/log"
	"github.com/qdm12/golibs/params"
	"github.com/qdm12/golibs/params/mock_params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getLogSettings(t *testing.T) {
	t.Parallel()

	var errDummy = errors.New("dummy")

	type onOffResult struct {
		on  bool
		err error
	}

	testCases := map[string]struct {
		logRequests  onOffResult
		logResponses onOffResult
		settings     log.Settings
		err          error
	}{
		"defaults": {},
		"log requests error": {
			logRequests: onOffResult{err: errDummy},
			err:         errDummy,
		},
		"log responses error": {
			logResponses: onOffResult{err: errDummy},
			err:          errDummy,
		},
		"all enabled": {
			logRequests:  onOffResult{on: true},
			logResponses: onOffResult{on: true},
			settings: log.Settings{
				LogRequests:  true,
				LogResponses: true,
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
				OnOff("LOG_REQUESTS", assignableDefault).
				Return(testCase.logRequests.on, testCase.logRequests.err)

			if testCase.logRequests.err == nil {
				env.EXPECT().
					OnOff("LOG_RESPONSES", assignableDefault).
					Return(testCase.logResponses.on, testCase.logResponses.err)
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
