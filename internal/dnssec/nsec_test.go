package dnssec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_nsecCover(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		zone      string
		nsecOwner string
		nsecNext  string
		ok        bool
	}{
		"zone_shorter_than_owner": {
			zone:      "example.com.",
			nsecOwner: "a.example.com.",
		},
		"zone_before_owner": {
			zone:      "a.example.com.",
			nsecOwner: "b.example.com.",
		},
		"zone_not_subdomain_of_owner": {
			zone:      "a.a.example.com.",
			nsecOwner: "b.example.com.",
		},
		"malformed_longer_next": {
			zone:      "b.example.com.",
			nsecOwner: "a.example.com.",
			nsecNext:  "c.c.example.com.",
		},
		"zone_equal_to_next": {
			zone:      "b.example.com.",
			nsecOwner: "a.example.com.",
			nsecNext:  "b.example.com.",
		},
		"zone_after_next": {
			zone:      "c.example.com.",
			nsecOwner: "a.example.com.",
			nsecNext:  "b.example.com.",
		},
		"zone_not_subdomain_of_next": {
			zone:      "b.b.example.com.",
			nsecOwner: "a.example.com.",
			nsecNext:  "c.example.com.",
			ok:        true,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ok := nsecCoversZone(testCase.zone, testCase.nsecOwner, testCase.nsecNext)

			assert.Equal(t, testCase.ok, ok)
		})
	}
}
