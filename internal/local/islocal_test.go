package local

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsFQDNLocal(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fqdn    string
		isLocal bool
	}{
		"no_dot": {
			fqdn:    "localhost.",
			isLocal: true,
		},
		"common_local_tld": {
			fqdn:    "x.lan.",
			isLocal: true,
		},
		"non_existing_tld": {
			fqdn:    "x.nonexisting.",
			isLocal: true,
		},
		"icann_managed_com": {
			fqdn: "x.com.",
		},
		"icann_managed_co_uk": {
			fqdn: "x.co.uk.",
		},
		"icann_managed_org": {
			fqdn: "x.y.org.",
		},
		"dyndns_privately_managed": {
			fqdn: "x.dyndns.org.",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			isLocal := IsFQDNLocal(testCase.fqdn)

			assert.Equal(t, testCase.isLocal, isLocal)
		})
	}
}
