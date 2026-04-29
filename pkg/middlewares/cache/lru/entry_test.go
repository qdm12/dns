package lru

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func Test_getExpUnix(t *testing.T) {
	t.Parallel()

	const nowUnix = int64(1000000000)

	testCases := map[string]struct {
		response *dns.Msg
		wantExp  int64
	}{
		"with_answer_ttl": {
			response: &dns.Msg{
				Answer: []dns.RR{
					&dns.A{
						Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600},
						A:   []byte{192, 0, 2, 1},
					},
				},
			},
			wantExp: 1000003600,
		},
		"with_multiple_answers_uses_minimum_ttl": {
			response: &dns.Msg{
				Answer: []dns.RR{
					&dns.A{
						Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600},
						A:   []byte{192, 0, 2, 1},
					},
					&dns.A{
						Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
						A:   []byte{192, 0, 2, 2},
					},
				},
			},
			wantExp: nowUnix + 300,
		},
		"with_rrsig_expiration": {
			response: &dns.Msg{
				Answer: []dns.RR{
					&dns.A{
						Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600},
						A:   []byte{192, 0, 2, 1},
					},
					&dns.RRSIG{
						Hdr:         dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeRRSIG, Class: dns.ClassINET, Ttl: 300},
						TypeCovered: dns.TypeA,
						Algorithm:   dns.ED25519,
						Labels:      2,
						Expiration:  1000000600, // expires in 600 seconds from now
					},
				},
			},
			// Should use RRSIG expiration (600 seconds) instead of record TTL (3600 seconds).
			wantExp: 1000000600,
		},
		"with_nsec_in_authority": {
			response: &dns.Msg{
				Answer: []dns.RR{
					&dns.A{
						Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600},
						A:   []byte{192, 0, 2, 1},
					},
				},
				Ns: []dns.RR{
					&dns.NSEC{
						Hdr:        dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeNSEC, Class: dns.ClassINET, Ttl: 300},
						NextDomain: "other.com.",
						TypeBitMap: []uint16{dns.TypeA, dns.TypeRRSIG, dns.TypeNSEC},
					},
				},
			},
			// Should use minimum of answer TTL (3600) and NSEC TTL (300).
			wantExp: 1000000300,
		},
		"with_soa_minimum": {
			response: &dns.Msg{
				Ns: []dns.RR{
					&dns.SOA{
						Hdr:     dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 86400},
						Ns:      "ns.example.com.",
						Mbox:    "admin.example.com.",
						Serial:  2024042701,
						Refresh: 10800,
						Retry:   3600,
						Expire:  604800,
						Minttl:  900,
					},
				},
			},
			// Should use minimum of SOA TTL (86400) and SOA MINIMUM (900).
			wantExp: 1000000900,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			expUnix := getExpUnix(testCase.response, nowUnix)
			assert.Equal(t, testCase.wantExp, expUnix)
		})
	}
}

func Test_verifyRRSIGValidity(t *testing.T) {
	t.Parallel()

	const nowUnix = int64(1000000000)

	testCases := map[string]struct {
		response *dns.Msg
		want     bool
	}{
		"valid": {
			response: &dns.Msg{
				Answer: []dns.RR{
					&dns.A{
						Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600},
						A:   []byte{192, 0, 2, 1},
					},
					&dns.RRSIG{
						Hdr:         dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeRRSIG, Class: dns.ClassINET, Ttl: 300},
						TypeCovered: dns.TypeA,
						Algorithm:   dns.ED25519,
						Labels:      2,
						Expiration:  uint32(nowUnix + 3600), // expires in 1 hour
					},
				},
			},
			want: true,
		},
		"expired": {
			response: &dns.Msg{
				Answer: []dns.RR{
					&dns.A{
						Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600},
						A:   []byte{192, 0, 2, 1},
					},
					&dns.RRSIG{
						Hdr:         dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeRRSIG, Class: dns.ClassINET, Ttl: 300},
						TypeCovered: dns.TypeA,
						Algorithm:   dns.ED25519,
						Labels:      2,
						Expiration:  uint32(nowUnix - 3600), // expired 1 hour ago
					},
				},
			},
			want: false,
		},
		"no_rrsig": {
			response: &dns.Msg{
				Answer: []dns.RR{
					&dns.A{
						Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600},
						A:   []byte{192, 0, 2, 1},
					},
				},
			},
			want: true,
		},
		"rrsig_in_authority": {
			response: &dns.Msg{
				Ns: []dns.RR{
					&dns.RRSIG{
						Hdr:         dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeRRSIG, Class: dns.ClassINET, Ttl: 300},
						TypeCovered: dns.TypeNSEC,
						Algorithm:   dns.ED25519,
						Labels:      2,
						Expiration:  uint32(nowUnix - 3600), // expired 1 hour ago
					},
				},
			},
			want: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			valid := verifyRRSIGValidity(testCase.response, nowUnix)
			assert.Equal(t, testCase.want, valid)
		})
	}
}
