package services

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/qdm12/dns/v2/internal/services/hooks"
	"github.com/stretchr/testify/assert"
)

func Test_SequenceSettings_SetDefaults(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		originalSettings  SequenceSettings
		defaultedSettings SequenceSettings
	}{
		"empty settings": {
			defaultedSettings: SequenceSettings{
				Hooks: hooks.NewNoop(),
			},
		},
		"hooks already set": {
			originalSettings: SequenceSettings{
				Hooks: hooks.NewWithLog(nil),
			},
			defaultedSettings: SequenceSettings{
				Hooks: hooks.NewWithLog(nil),
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testCase.originalSettings.SetDefaults()
			assert.Equal(t, testCase.defaultedSettings, testCase.originalSettings)
		})
	}
}

func Test_SequenceSettings_Validate(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	// Need to share the same service pointers so they are defined in the
	// parent test for all the subtests.
	dummyServiceOne := NewMockService(ctrl)
	dummyServiceOne.EXPECT().String().Return("dummy one").AnyTimes()
	dummyServiceTwo := NewMockService(ctrl)
	dummyServiceTwo.EXPECT().String().Return("dummy two").AnyTimes()

	testCases := map[string]struct {
		settings    SequenceSettings
		errSentinel error
		errMessage  string
	}{
		"no service start order": {
			settings:    SequenceSettings{},
			errSentinel: ErrNoServiceStart,
			errMessage:  "no service start order specified",
		},
		"no service stop order": {
			settings: SequenceSettings{
				ServicesStart: []Service{dummyServiceOne},
			},
			errSentinel: ErrNoServiceStop,
			errMessage:  "no service stop order specified",
		},
		"service to start and stop size mismatch": {
			settings: SequenceSettings{
				ServicesStart: []Service{dummyServiceOne, dummyServiceTwo},
				ServicesStop:  []Service{dummyServiceOne},
			},
			errSentinel: ErrServicesStartStopMismatch,
			errMessage: "services to start and stop mismatch: " +
				"2 services to start (dummy one and dummy two) and " +
				"1 services to stop (dummy one)",
		},
		"nil start service": {
			settings: SequenceSettings{
				ServicesStart: []Service{nil},
				ServicesStop:  []Service{dummyServiceOne},
			},
			errSentinel: ErrServiceIsNil,
			errMessage:  "service to start at index 0: service is nil",
		},
		"nil stop service": {
			settings: SequenceSettings{
				ServicesStart: []Service{dummyServiceOne},
				ServicesStop:  []Service{nil},
			},
			errSentinel: ErrServiceIsNil,
			errMessage:  "service to stop at index 0: service is nil",
		},
		"start and stop service do not match": {
			settings: SequenceSettings{
				ServicesStart: []Service{dummyServiceOne},
				ServicesStop:  []Service{dummyServiceTwo},
			},
			errSentinel: ErrServicesStartStopMismatch,
			errMessage: "services to start and stop mismatch: " +
				"service to start dummy one is not the service to stop dummy two",
		},
		"start and stop services do not match": {
			settings: SequenceSettings{
				ServicesStart: []Service{dummyServiceOne, dummyServiceOne},
				ServicesStop:  []Service{dummyServiceTwo, dummyServiceTwo},
			},
			errSentinel: ErrServicesStartStopMismatch,
			errMessage: "services to start and stop mismatch: " +
				"services to start dummy one and dummy one are not " +
				"the services to stop dummy two and dummy two",
		},
		"single service duplicated": {
			settings: SequenceSettings{
				ServicesStart: []Service{dummyServiceOne, dummyServiceOne, dummyServiceOne},
				ServicesStop:  []Service{dummyServiceOne, dummyServiceOne, dummyServiceOne},
			},
			errSentinel: ErrServicesNotUnique,
			errMessage:  "services are not unique: service dummy one is duplicated 3 times",
		},
		"multiple services duplicated": {
			settings: SequenceSettings{
				ServicesStart: []Service{dummyServiceOne, dummyServiceOne, dummyServiceTwo, dummyServiceTwo},
				ServicesStop:  []Service{dummyServiceTwo, dummyServiceTwo, dummyServiceOne, dummyServiceOne},
			},
			errSentinel: ErrServicesNotUnique,
			errMessage: "services are not unique: services dummy one is duplicated twice " +
				"and dummy two is duplicated twice",
		},
		"success": {
			settings: SequenceSettings{
				ServicesStart: []Service{dummyServiceOne},
				ServicesStop:  []Service{dummyServiceOne},
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := testCase.settings.Validate()

			assert.ErrorIs(t, err, testCase.errSentinel)
			if testCase.errSentinel != nil {
				assert.EqualError(t, err, testCase.errMessage)
			}
		})
	}
}
