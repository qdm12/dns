package services

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func checkErrIsErrTest(t *testing.T, err error, serviceName string,
	sentinelErr error) {
	t.Helper()
	assert.ErrorIs(t, err, sentinelErr)
	assert.EqualError(t, err, serviceName+" crashed: "+sentinelErr.Error())
	expectedServiceErr := serviceError{
		format:      errorFormatCrash,
		serviceName: serviceName,
		err:         sentinelErr,
	}
	assert.Equal(t, expectedServiceErr, err)
}

func Test_andStrings(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		strings []string
		result  string
	}{
		"empty": {},
		"single": {
			strings: []string{"A"},
			result:  "A",
		},
		"two": {
			strings: []string{"A", "B"},
			result:  "A and B",
		},
		"three": {
			strings: []string{"A", "B", "C"},
			result:  "A, B and C",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := andStrings(testCase.strings)
			assert.Equal(t, testCase.result, result)
		})
	}
}

func Test_andServiceStrings(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		servicesBuilder func(ctrl *gomock.Controller) []Service
		result          string
	}{
		"empty": {
			servicesBuilder: func(ctrl *gomock.Controller) []Service {
				return nil
			},
		},
		"single": {
			servicesBuilder: func(ctrl *gomock.Controller) []Service {
				serviceA := NewMockService(ctrl)
				serviceA.EXPECT().String().Return("A")
				return []Service{serviceA}
			},
			result: "A",
		},
		"two": {
			servicesBuilder: func(ctrl *gomock.Controller) []Service {
				serviceA := NewMockService(ctrl)
				serviceA.EXPECT().String().Return("A")
				serviceB := NewMockService(ctrl)
				serviceB.EXPECT().String().Return("B")
				return []Service{serviceA, serviceB}
			},
			result: "A and B",
		},
		"three": {
			servicesBuilder: func(ctrl *gomock.Controller) []Service {
				serviceA := NewMockService(ctrl)
				serviceA.EXPECT().String().Return("A")
				serviceB := NewMockService(ctrl)
				serviceB.EXPECT().String().Return("B")
				serviceC := NewMockService(ctrl)
				serviceC.EXPECT().String().Return("C")
				return []Service{serviceA, serviceB, serviceC}
			},
			result: "A, B and C",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			services := testCase.servicesBuilder(ctrl)

			result := andServiceStrings(services)
			assert.Equal(t, testCase.result, result)
		})
	}
}
