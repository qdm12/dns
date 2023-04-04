package services

import (
	"testing"
	"time"

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

type syncMutexTest interface {
	TryLock() bool
	Unlock()
}

func assertMutexUnlocked(t *testing.T, mutex syncMutexTest) {
	t.Helper()
	if !mutex.TryLock() {
		t.Error("mutex is locked")
	}
	mutex.Unlock()
}

func assertNoRunError(t *testing.T, runError <-chan error) {
	t.Helper()
	select {
	case err := <-runError:
		t.Errorf("unexpected run error: %s", err)
	default:
	}
}

func assertRunError(t *testing.T, runError <-chan error, expectedErr error) {
	t.Helper()
	const timeout = 100 * time.Millisecond
	timer := time.NewTimer(timeout)
	select {
	case err := <-runError:
		assert.ErrorIs(t, err, expectedErr)
	case <-timer.C:
		t.Fatal("run error not received")
	}

	timer.Stop()
	timer.Reset(timeout)
	select {
	case <-runError:
	case <-timer.C:
		t.Fatal("run error not closed after receiving error")
	}
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
