package dnssec

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func Test_groupRRs(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		rrs          []dns.RR
		dnssecRRSets []dnssecRRSet
		errWrapped   error
		errMessage   string
	}{
		"no_rrs": {
			dnssecRRSets: []dnssecRRSet{},
		},
		"bad_single_rrsig_answer": {
			rrs: []dns.RR{
				newEmptyRRSig(dns.TypeA),
			},
			errWrapped: ErrRRSigForNoRRSet,
			errMessage: "for RRSet example.com. A: RRSIG for no RRSet",
		},
		"bad_rrsig_for_no_rrset": {
			rrs: []dns.RR{
				newEmptyAAAA(),
				newEmptyRRSig(dns.TypeAAAA),
				newEmptyRRSig(dns.TypeA), // bad one
			},
			errWrapped: ErrRRSigForNoRRSet,
			errMessage: "for RRSet example.com. A: RRSIG for no RRSet",
		},
		"multiple_rrsig_for_same_type": {
			rrs: []dns.RR{
				newEmptyRRSig(dns.TypeA),
				newEmptyA(),
				newEmptyRRSig(dns.TypeA),
			},
			dnssecRRSets: []dnssecRRSet{
				{
					rrSet: []dns.RR{
						newEmptyA(),
					},
					rrSigs: []*dns.RRSIG{
						newEmptyRRSig(dns.TypeA),
						newEmptyRRSig(dns.TypeA),
					},
				},
			},
		},
		"bad_signed_and_not_signed_rrsets": {
			rrs: []dns.RR{
				newEmptyRRSig(dns.TypeA),
				newEmptyAAAA(),
				newEmptyA(),
			},
			errWrapped: ErrRRSetSignedAndUnsigned,
			errMessage: "mix of signed and unsigned RRSets: 1 signed and 1 unsigned RRSets",
		},
		"signed_rrsets": {
			rrs: []dns.RR{
				newEmptyRRSig(dns.TypeA),
				newEmptyA(),
				newEmptyAAAA(),
				newEmptyRRSig(dns.TypeAAAA),
			},
			dnssecRRSets: []dnssecRRSet{
				{
					rrSigs: []*dns.RRSIG{newEmptyRRSig(dns.TypeA)},
					rrSet: []dns.RR{
						newEmptyA(),
					},
				},
				{
					rrSigs: []*dns.RRSIG{newEmptyRRSig(dns.TypeAAAA)},
					rrSet: []dns.RR{
						newEmptyAAAA(),
					},
				},
			},
		},
		"not_signed_rrsets": {
			rrs: []dns.RR{
				newEmptyA(),
				newEmptyAAAA(),
			},
			dnssecRRSets: []dnssecRRSet{
				{
					rrSet: []dns.RR{
						newEmptyA(),
					},
				},
				{
					rrSet: []dns.RR{
						newEmptyAAAA(),
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dnssecRRSets, err := groupRRs(testCase.rrs)

			assert.Equal(t, testCase.dnssecRRSets, dnssecRRSets)
			assert.ErrorIs(t, err, testCase.errWrapped)
			if testCase.errWrapped != nil {
				assert.EqualError(t, err, testCase.errMessage)
			}
		})
	}
}

func newEmptyRRSig(typeCovered uint16) *dns.RRSIG {
	return &dns.RRSIG{
		Hdr: dns.RR_Header{
			Name:   "example.com.",
			Rrtype: dns.TypeRRSIG,
		},
		TypeCovered: typeCovered,
		SignerName:  "example.com.",
	}
}

func newEmptyA() *dns.A {
	return &dns.A{
		Hdr: dns.RR_Header{
			Name:   "example.com.",
			Rrtype: dns.TypeA,
		},
	}
}

func newEmptyAAAA() *dns.AAAA {
	return &dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   "example.com.",
			Rrtype: dns.TypeAAAA,
		},
	}
}
