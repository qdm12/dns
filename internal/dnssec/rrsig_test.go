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

func Test_checkRRSigSignerName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		rrSig           *dns.RRSIG
		keyTagToDNSKeys dnsKeysByTag
		errWrapped      error
		errMessage      string
	}{
		"signer_matches_dnskey_owner": {
			rrSig: &dns.RRSIG{
				Hdr: dns.RR_Header{
					Name: "example.com.",
				},
				TypeCovered: dns.TypeA,
				SignerName:  "example.com.",
				KeyTag:      42,
			},
			keyTagToDNSKeys: dnsKeysByTag{
				42: {&dns.DNSKEY{Hdr: dns.RR_Header{Name: "example.com."}}},
			},
		},
		"signer_mismatch": {
			rrSig: &dns.RRSIG{
				Hdr: dns.RR_Header{
					Name: "example.com.",
				},
				TypeCovered: dns.TypeA,
				SignerName:  "com.",
				KeyTag:      42,
			},
			keyTagToDNSKeys: dnsKeysByTag{
				42: {&dns.DNSKEY{Hdr: dns.RR_Header{Name: "example.com."}}},
			},
			errWrapped: errRRSigSignerName,
			errMessage: `for RRSIG for owner example.com. and type A: ` +
				`RRSIG signer name is not zone apex: "com." should be "example.com."`,
		},
		"key_tag_not_found": {
			rrSig: &dns.RRSIG{
				Hdr: dns.RR_Header{
					Name: "example.com.",
				},
				TypeCovered: dns.TypeA,
				SignerName:  "example.com.",
				KeyTag:      99,
			},
			keyTagToDNSKeys: dnsKeysByTag{},
			errWrapped:      errRRSigDNSKey,
			errMessage:      `DNSKEY not found: for key tag 99`,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := checkRRSigSignerName(testCase.rrSig, testCase.keyTagToDNSKeys)

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
		Algorithm:  dns.ED25519,
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
		"forbidden_algorithm": {
			rrSigs:     []*dns.RRSIG{{Algorithm: dns.RSAMD5}},
			budget:     newRRSIGValidationBudget(),
			errWrapped: errRRSigForbiddenAlgorithm,
		},
		"unsupported_algorithm": {
			rrSigs:     []*dns.RRSIG{{Algorithm: 255}},
			budget:     newRRSIGValidationBudget(),
			errWrapped: errRRSigUnsupportedAlgorithm,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := verifyRRSetRRSigs(rrSet, testCase.rrSigs, dnsKeysByTag{}, testCase.budget)

			require.Error(t, err)
			assert.ErrorIs(t, err, testCase.errWrapped)
		})
	}
}
