package services

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_validateServicesAreUnique(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	serviceA := NewMockService(ctrl)
	serviceA.EXPECT().String().Return("A").AnyTimes()

	serviceA2 := NewMockService(ctrl)
	serviceA2.EXPECT().String().Return("A").AnyTimes()

	serviceB := NewMockService(ctrl)
	serviceB.EXPECT().String().Return("B").AnyTimes()

	testCases := map[string]struct {
		services   []Service
		errMessage string
	}{
		"no service": {},
		"single service": {
			services: []Service{serviceA},
		},
		"two same services": {
			services:   []Service{serviceA, serviceA},
			errMessage: "service A is duplicated twice",
		},
		"two same string services": {
			services:   []Service{serviceA, serviceA2},
			errMessage: "service name A is duplicated twice",
		},
		"two distinct services": {
			services: []Service{serviceA, serviceB},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			errMessage := validateServicesAreUnique(testCase.services)
			assert.Equal(t, testCase.errMessage, errMessage)
		})
	}
}

func Test_findDuplicateServices(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	serviceA := NewMockService(ctrl)
	serviceA.EXPECT().String().Return("A").AnyTimes()

	serviceA2 := NewMockService(ctrl)
	serviceA2.EXPECT().String().Return("A").AnyTimes()

	serviceB := NewMockService(ctrl)
	serviceB.EXPECT().String().Return("B").AnyTimes()

	testCases := map[string]struct {
		services          []Service
		duplicateServices map[fmt.Stringer]uint
		duplicatedNames   map[string]uint
	}{
		"no service": {
			duplicateServices: map[fmt.Stringer]uint{},
			duplicatedNames:   map[string]uint{},
		},
		"single service": {
			services:          []Service{serviceA},
			duplicateServices: map[fmt.Stringer]uint{},
			duplicatedNames:   map[string]uint{},
		},
		"two same services": {
			services: []Service{serviceA, serviceA},
			duplicateServices: map[fmt.Stringer]uint{
				serviceA: 2,
			},
			duplicatedNames: map[string]uint{
				"A": 2,
			},
		},
		"two same string services": {
			services:          []Service{serviceA, serviceA2},
			duplicateServices: map[fmt.Stringer]uint{},
			duplicatedNames: map[string]uint{
				"A": 2,
			},
		},
		"two distinct services": {
			services:          []Service{serviceA, serviceB},
			duplicateServices: map[fmt.Stringer]uint{},
			duplicatedNames:   map[string]uint{},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			duplicateServices, duplicatedNames := findDuplicatedServices(testCase.services)
			assert.Equal(t, testCase.duplicateServices, duplicateServices)
			assert.Equal(t, testCase.duplicatedNames, duplicatedNames)
		})
	}
}

func Test_makeDuplicatedServicesErrMessage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	serviceA := NewMockService(ctrl)
	serviceA.EXPECT().String().Return("A").AnyTimes()

	serviceB := NewMockService(ctrl)
	serviceB.EXPECT().String().Return("B").AnyTimes()

	testCases := map[string]struct {
		duplicateServices map[fmt.Stringer]uint
		messagePrefix     string
		errMessage        string
	}{
		"no duplicated service": {
			duplicateServices: map[fmt.Stringer]uint{},
		},
		"single service duplicated 3 times": {
			duplicateServices: map[fmt.Stringer]uint{
				serviceA: 3,
			},
			messagePrefix: "service",
			errMessage:    "service A is duplicated 3 times",
		},
		"two services duplicated twice": {
			duplicateServices: map[fmt.Stringer]uint{
				serviceA: 2,
				serviceB: 2,
			},
			messagePrefix: "service",
			errMessage: "services A is duplicated twice and " +
				"B is duplicated twice",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			errMessage := makeDuplicatedServicesErrMessage(
				testCase.duplicateServices, testCase.messagePrefix)
			assert.Equal(t, testCase.errMessage, errMessage)
		})
	}
}

func Test_countToString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		count uint
		s     string
	}{
		"zero": {
			s: "0 time",
		},
		"one": {
			count: 1,
			s:     "once",
		},
		"two": {
			count: 2,
			s:     "twice",
		},
		"three": {
			count: 3,
			s:     "3 times",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := countToString(testCase.count)
			assert.Equal(t, testCase.s, s)
		})
	}
}
