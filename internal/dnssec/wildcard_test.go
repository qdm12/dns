package dnssec

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func Test_answerHasWildcardedDNAME(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		answerRRSets  []dnssecRRSet
		hasWildcarded bool
	}{
		"empty_answer": {
			answerRRSets:  []dnssecRRSet{},
			hasWildcarded: false,
		},
		"answer_with_a_records": {
			answerRRSets: []dnssecRRSet{
				{
					rrSet: []dns.RR{
						&dns.A{
							Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET},
							A:   []byte{192, 0, 2, 1},
						},
					},
				},
			},
			hasWildcarded: false,
		},
		"answer_with_dname": {
			answerRRSets: []dnssecRRSet{
				{
					rrSet: []dns.RR{
						&dns.DNAME{
							Hdr:    dns.RR_Header{Name: "*.example.com.", Rrtype: dns.TypeDNAME, Class: dns.ClassINET},
							Target: "other.com.",
						},
					},
				},
			},
			hasWildcarded: true,
		},
		"answer_with_multiple_records_including_dname": {
			answerRRSets: []dnssecRRSet{
				{
					rrSet: []dns.RR{
						&dns.A{
							Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET},
							A:   []byte{192, 0, 2, 1},
						},
					},
				},
				{
					rrSet: []dns.RR{
						&dns.DNAME{
							Hdr:    dns.RR_Header{Name: "*.example.com.", Rrtype: dns.TypeDNAME, Class: dns.ClassINET},
							Target: "other.com.",
						},
					},
				},
			},
			hasWildcarded: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			hasWildcarded := answerHasWildcardedDNAME(testCase.answerRRSets)

			assert.Equal(t, testCase.hasWildcarded, hasWildcarded)
		})
	}
}
