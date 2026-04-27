package dnssec

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			errWrapped: errRRSigSignerName,
			errMessage: `for RRSIG for owner example.com. and type A: ` +
				`signer name is not valid: "." should be "example.com." or "com."`,
		},
		"root_dnskey_signer_is_root": {
			rrSig: &dns.RRSIG{
				Hdr: dns.RR_Header{
					Name: ".",
				},
				TypeCovered: dns.TypeDNSKEY,
				SignerName:  ".",
			},
		},
		"root_dnskey_signer_is_invalid": {
			rrSig: &dns.RRSIG{
				Hdr: dns.RR_Header{
					Name: ".",
				},
				TypeCovered: dns.TypeDNSKEY,
				SignerName:  "com.",
			},
			errWrapped: errRRSigSignerName,
			errMessage: `for RRSIG for owner . and type DNSKEY: ` +
				`signer name is not valid: "com." should be "."`,
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
			errWrapped: errRRSigSignerName,
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
			errWrapped: errRRSigSignerName,
			errMessage: `for RRSIG for owner example.com. and type CNAME: ` +
				`signer name is not valid: "example.com." should be "com."`,
		},
	}

	for name, testCase := range testCases {
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

func Test_verifyRRSetRRSigs(t *testing.T) {
	t.Parallel()

	rrSet := []dns.RR{
		&dns.A{Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET}},
	}

	const now = uint32(1000)
	overBudgetRRSigs := make([]*dns.RRSIG, maxRRSIGValidationsPerRRSet+1)
	for i := range overBudgetRRSigs {
		overBudgetRRSigs[i] = &dns.RRSIG{}
	}
	singleRRSig := []*dns.RRSIG{{
		Hdr:        dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeRRSIG, Class: dns.ClassINET},
		KeyTag:     12345,
		Expiration: now + 60,
		Inception:  now - 60,
	}}

	testCases := map[string]struct {
		rrSigs     []*dns.RRSIG
		budget     *rrsigValidationBudget
		errWrapped error
	}{
		"rrset_budget_exceeded": {
			rrSigs:     overBudgetRRSigs,
			budget:     newRRSIGValidationBudget(),
			errWrapped: errRRSIGValidationRRSetBudgetExceeded,
		},
		"message_budget_exceeded": {
			rrSigs:     singleRRSig,
			budget:     &rrsigValidationBudget{remaining: 0},
			errWrapped: errRRSIGValidationBudgetExceeded,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := verifyRRSetRRSigs(rrSet, testCase.rrSigs, map[uint16]*dns.DNSKEY{}, testCase.budget)

			require.Error(t, err)
			assert.ErrorIs(t, err, testCase.errWrapped)
		})
	}
}
