package dnssec

import (
	"testing"

	"github.com/miekg/dns"
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
			ok:        true,
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
		"wildcard_a": {
			zone:      "a.example.com.",
			nsecOwner: "*.example.com.",
			nsecNext:  "example.com.",
			ok:        true,
		},
		"wildcard_a.a": {
			zone:      "a.a.example.com.",
			nsecOwner: "*.example.com.",
			nsecNext:  "example.com.",
			ok:        true,
		},
		"wildcard_#": {
			zone:      "#.example.com.",
			nsecOwner: "*.example.com.",
			nsecNext:  "example.com.",
			ok:        false,
		},
		"wrapped_interval_after_owner": {
			zone:      "z.example.com.",
			nsecOwner: "y.example.com.",
			nsecNext:  "b.example.com.",
			ok:        true,
		},
		"wrapped_interval_before_next": {
			zone:      "a.example.com.",
			nsecOwner: "y.example.com.",
			nsecNext:  "b.example.com.",
			ok:        true,
		},
		"wrapped_interval_middle_gap": {
			zone:      "c.example.com.",
			nsecOwner: "y.example.com.",
			nsecNext:  "b.example.com.",
			ok:        false,
		},
		"canonical_label_length_shorter_first": {
			zone:      "ylyly.example.com.",
			nsecOwner: "x.example.com.",
			nsecNext:  "z.example.com.",
			ok:        true,
		},
		"canonical_label_length_not_after_z": {
			zone:      "z-zzz.example.com.",
			nsecOwner: "x.example.com.",
			nsecNext:  "z.example.com.",
			ok:        false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ok := nsecCoversZone(testCase.zone, testCase.nsecOwner, testCase.nsecNext)

			assert.Equal(t, testCase.ok, ok)
		})
	}
}

func Test_canonicalNameCompare(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		a        string
		b        string
		expected int
	}{
		"equal_case_insensitive": {
			a:        "A.Example.COM.",
			b:        "a.example.com.",
			expected: 0,
		},
		"shorter_label_first": {
			a:        "z.example.com.",
			b:        "z-zzz.example.com.",
			expected: -1,
		},
		"from_rfc_comment_ylyly_before_z": {
			a:        "ylyly.example.com.",
			b:        "z.example.com.",
			expected: -1,
		},
		"more_labels_after_if_suffix_equal": {
			a:        "a.example.com.",
			b:        "example.com.",
			expected: 1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := canonicalNameCompare(tc.a, tc.b)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func Test_findNoDataDSProofNSEC_prefers_exact_match(t *testing.T) {
	t.Parallel()

	const qname = "f.ip6.arpa."

	// First NSEC covers qname by interval but is not an exact owner match.
	coveringNSEC := &dns.NSEC{
		Hdr: dns.RR_Header{
			Name:   "e.ip6.arpa.",
			Rrtype: dns.TypeNSEC,
			Class:  dns.ClassINET,
		},
		NextDomain: "g.ip6.arpa.",
		TypeBitMap: []uint16{dns.TypeDS, dns.TypeRRSIG, dns.TypeNSEC},
	}

	// Second NSEC is the correct exact owner-name proof for qname.
	exactNSEC := &dns.NSEC{
		Hdr: dns.RR_Header{
			Name:   qname,
			Rrtype: dns.TypeNSEC,
			Class:  dns.ClassINET,
		},
		NextDomain: "z.ip6.arpa.",
		TypeBitMap: []uint16{dns.TypeNS, dns.TypeRRSIG, dns.TypeNSEC},
	}

	exactMatch, covering := findNoDataDSProofNSEC(qname, []dns.RR{coveringNSEC, exactNSEC})

	assert.Same(t, exactNSEC, exactMatch)
	assert.Same(t, coveringNSEC, covering)
}

func Test_nsecValidateNoDataDS_accepts_covering_nsec(t *testing.T) {
	t.Parallel()

	err := nsecValidateNoDataDS("0.224.in-addr.arpa.", []dns.RR{
		&dns.NSEC{
			Hdr: dns.RR_Header{
				Name:   "224.in-addr.arpa.",
				Rrtype: dns.TypeNSEC,
				Class:  dns.ClassINET,
			},
			NextDomain: "0.0.0.224.in-addr.arpa.",
			TypeBitMap: []uint16{dns.TypeNS, dns.TypeSOA, dns.TypeRRSIG, dns.TypeNSEC, dns.TypeDNSKEY},
		},
	})

	assert.NoError(t, err)
}
