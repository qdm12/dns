package dnssec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getNextCloser(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		qname           string
		closestEncloser string
		nextCloser      string
	}{
		"case1": {
			qname:           "a.b.example.com.",
			closestEncloser: "example.com.",
			nextCloser:      "b.example.com.",
		},
		"q_name_is_next_closer": {
			qname:           "a.example.com.",
			closestEncloser: "example.com.",
			nextCloser:      "a.example.com.",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			nextCloser := getNextCloser(testCase.qname, testCase.closestEncloser)

			assert.Equal(t, testCase.nextCloser, nextCloser)
		})
	}
}
