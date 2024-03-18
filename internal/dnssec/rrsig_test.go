package dnssec

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func Test_sortRRSIGsByAlgo(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		rrSigs   []*dns.RRSIG
		expected []*dns.RRSIG
	}{
		"empty": {},
		"single": {
			rrSigs: []*dns.RRSIG{
				{Algorithm: dns.RSASHA1},
			},
			expected: []*dns.RRSIG{
				{Algorithm: dns.RSASHA1},
			},
		},
		"multiple": {
			rrSigs: []*dns.RRSIG{
				{Algorithm: dns.ED25519},
				{Algorithm: dns.RSASHA1},
				{Algorithm: dns.ECCGOST},
				{Algorithm: dns.RSASHA512},
				{Algorithm: dns.ECDSAP384SHA384},
				{Algorithm: dns.DSA},
			},
			expected: []*dns.RRSIG{
				{Algorithm: dns.ED25519},
				{Algorithm: dns.ECDSAP384SHA384},
				{Algorithm: dns.RSASHA1},
				{Algorithm: dns.RSASHA512},
				{Algorithm: dns.ECCGOST},
				{Algorithm: dns.DSA},
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			sortRRSIGsByAlgo(testCase.rrSigs)

			assert.Equal(t, testCase.expected, testCase.rrSigs)
		})
	}
}

func Test_rrSigCheckSignerName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		rrSig      *dns.RRSIG
		errWrapped error
		errMessage string
	}{
		"a_signer_is_owner": {
			rrSig: &dns.RRSIG{
				Hdr: dns.RR_Header{
					Name: "example.com.",
				},
				TypeCovered: dns.TypeA,
				SignerName:  "example.com.",
			},
		},
		"a_signer_is_parent": {
			rrSig: &dns.RRSIG{
				Hdr: dns.RR_Header{
					Name: "example.com.",
				},
				TypeCovered: dns.TypeA,
				SignerName:  "com.",
			},
		},
		"a_signer_is_invalid": {
			rrSig: &dns.RRSIG{
				Hdr: dns.RR_Header{
					Name: "example.com.",
				},
				TypeCovered: dns.TypeA,
				SignerName:  ".",
			},
			errWrapped: ErrRRSigSignerName,
			errMessage: `for RRSIG for owner example.com. and type A: ` +
				`signer name is not valid: "." should be "example.com." or "com."`,
		},
		"ds_signer_is_parent": {
			rrSig: &dns.RRSIG{
				Hdr: dns.RR_Header{
					Name: "example.com.",
				},
				TypeCovered: dns.TypeDS,
				SignerName:  "com.",
			},
		},
		"ds_signer_is_owner": {
			rrSig: &dns.RRSIG{
				Hdr: dns.RR_Header{
					Name: "example.com.",
				},
				TypeCovered: dns.TypeDS,
				SignerName:  "example.com.",
			},
			errWrapped: ErrRRSigSignerName,
			errMessage: `for RRSIG for owner example.com. and type DS: ` +
				`signer name is not valid: "example.com." should be "com."`,
		},
		"cname_signer_is_parent": {
			rrSig: &dns.RRSIG{
				Hdr: dns.RR_Header{
					Name: "example.com.",
				},
				TypeCovered: dns.TypeCNAME,
				SignerName:  "com.",
			},
		},
		"cname_signer_is_owner": {
			rrSig: &dns.RRSIG{
				Hdr: dns.RR_Header{
					Name: "example.com.",
				},
				TypeCovered: dns.TypeCNAME,
				SignerName:  "example.com.",
			},
			errWrapped: ErrRRSigSignerName,
			errMessage: `for RRSIG for owner example.com. and type CNAME: ` +
				`signer name is not valid: "example.com." should be "com."`,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := rrSigCheckSignerName(testCase.rrSig)

			assert.ErrorIs(t, err, testCase.errWrapped)
			if testCase.errWrapped != nil {
				assert.EqualError(t, err, testCase.errMessage)
			}
		})
	}
}
