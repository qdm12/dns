package local

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsFQDNLocal(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		publicNames []string
		fqdn        string
		isLocal     bool
	}{
		"no_dot": {
			fqdn:    "localhost.",
			isLocal: true,
		},
		"ipv4_literal": {
			fqdn: "192.168.1.1.",
		},
		"ipv6_literal": {
			fqdn: "::1.",
		},
		"ipv6_literal_bracketed": {
			fqdn: "[::1].",
		},
		"ipv6_unspecified_bracketed": {
			fqdn: "[::].",
		},
		"ipv6_global_bracketed": {
			fqdn: "[2001:db8::1].",
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
		"mixed_case": {
			fqdn: "weBsItE.Eu.oRG.",
		},
		"network_tld": {
			fqdn: "orpheus.network.",
		},
		"public_name_as_local": {
			publicNames: []string{"github.com"},
			fqdn:        "github.com.",
			isLocal:     true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			checker := New(testCase.publicNames)

			isLocal := checker.IsFQDNLocal(testCase.fqdn)

			assert.Equal(t, testCase.isLocal, isLocal)
		})
	}
}
