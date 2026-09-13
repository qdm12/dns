package dnssec

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getNextCloser(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		qname           string
		closestEncloser string
		nextCloser      string
	}{
		"case1": {
			qname:           "a.b.example.com.",
			closestEncloser: "example.com.",
			nextCloser:      "b.example.com.",
		},
		"q_name_is_next_closer": {
			qname:           "a.example.com.",
			closestEncloser: "example.com.",
			nextCloser:      "a.example.com.",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			nextCloser := getNextCloser(testCase.qname, testCase.closestEncloser)

			assert.Equal(t, testCase.nextCloser, nextCloser)
		})
	}
}

func newOisdNSEC3(owner, nextDomain string) *dns.NSEC3 {
	return &dns.NSEC3{
		Hdr:        dns.RR_Header{Name: owner, Rrtype: dns.TypeNSEC3, Class: dns.ClassINET},
		Hash:       dns.SHA1,
		Flags:      1,
		Iterations: 0,
		NextDomain: nextDomain,
		TypeBitMap: []uint16{dns.TypeA, dns.TypeAAAA, dns.TypeRRSIG},
	}
}

// newMatchingNSEC3 returns an NSEC3 RR of zone whose hashed owner name
// matches qname, computed with SHA1, no iterations and an empty salt.
func newMatchingNSEC3(qname, zone string, types ...uint16) *dns.NSEC3 {
	return &dns.NSEC3{
		Hdr: dns.RR_Header{
			Name:   dns.HashName(qname, dns.SHA1, 0, "") + "." + zone,
			Rrtype: dns.TypeNSEC3,
			Class:  dns.ClassINET,
		},
		Hash:       dns.SHA1,
		Flags:      0,
		Iterations: 0,
		TypeBitMap: types,
	}
}

func Test_nsec3ValidateNoDataDS(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		qname          string
		delegationName string
		types          []uint16
		errWrapped     error
	}{
		"delegation_name_with_ns": {
			qname:          "sub.example.com.",
			delegationName: "sub.example.com.",
			types:          []uint16{dns.TypeNS, dns.TypeA},
		},
		"delegation_name_without_ns": {
			qname:          "sub.example.com.",
			delegationName: "sub.example.com.",
			types:          []uint16{dns.TypeA},
			errWrapped:     errNSEC3NoDataDSNSNotSet,
		},
		"not_a_delegation_name_without_ns": {
			qname: "sub.example.com.",
			types: []uint16{dns.TypeA},
		},
		"other_delegation_name_without_ns": {
			qname:          "sub.example.com.",
			delegationName: "other.example.com.",
			types:          []uint16{dns.TypeA},
		},
		"delegation_name_with_ds_type": {
			qname:          "sub.example.com.",
			delegationName: "sub.example.com.",
			types:          []uint16{dns.TypeNS, dns.TypeDS},
			errWrapped:     errBogus,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			nsec3RRSet := []dns.RR{newMatchingNSEC3(testCase.qname, "example.com.",
				testCase.types...)}
			err := nsec3ValidateNoDataDS(testCase.qname, testCase.delegationName,
				nsec3RRSet)

			assert.ErrorIs(t, err, testCase.errWrapped)
		})
	}
}

func Test_nsec3ValidateWildcard(t *testing.T) {
	t.Parallel()

	// The NSEC3 RRs are the actual records returned by the oisd.nl.
	// zone, which uses SHA1 with 0 iterations and an empty salt, making
	// the hashes deterministic.
	testCases := map[string]struct {
		qname      string
		nsec3RRSet []dns.RR
		errWrapped error
	}{
		"covers_next_closer_small": {
			qname: "small.oisd.nl.",
			nsec3RRSet: []dns.RR{newOisdNSEC3("UU303NS3F3NROKRMNQPE83JE6AFONLA5.oisd.nl.",
				"0AEK26R7AMHP3KLCGN1R0F9IRN4SV46A.oisd.nl.")},
		},
		"covers_next_closer_big": {
			qname: "big.oisd.nl.",
			nsec3RRSet: []dns.RR{newOisdNSEC3("M614QHO57S65RJHQKG0G26L9IEFOSU2V.oisd.nl.",
				"TI4IMV94LMGCCTDEA183M3O2S1F7JG1I.oisd.nl.")},
		},
		"does_not_cover_next_closer": {
			qname: "big.oisd.nl.",
			nsec3RRSet: []dns.RR{newOisdNSEC3("UU303NS3F3NROKRMNQPE83JE6AFONLA5.oisd.nl.",
				"0AEK26R7AMHP3KLCGN1R0F9IRN4SV46A.oisd.nl.")},
			errWrapped: errBogus,
		},
		"no_nsec3_rrs": {
			qname:      "small.oisd.nl.",
			nsec3RRSet: []dns.RR{},
			errWrapped: errBogus,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := nsec3ValidateWildcard(testCase.qname, testCase.nsec3RRSet)

			assert.ErrorIs(t, err, testCase.errWrapped)
		})
	}
}

func Test_nsec3InitialChecks_IterationPolicy(t *testing.T) {
	t.Parallel()

	ed25519Key := &dns.DNSKEY{
		Hdr:       dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeDNSKEY, Class: dns.ClassINET},
		Flags:     dns.ZONE,
		Algorithm: dns.ED25519,
	}

	testCases := map[string]struct {
		iterations     uint16
		keyTagToDNSKey dnsKeysByTag
		errWrapped     error
		expectLen      int
	}{
		"rejects_iterations_above_absolute_cap_without_keys": {
			iterations: 2501,
			errWrapped: errNSEC3IterationsTooHigh,
		},
		"rejects_iterations_above_small_key_policy": {
			iterations:     151,
			keyTagToDNSKey: dnsKeysByTag{12345: {ed25519Key}},
			errWrapped:     errNSEC3IterationsTooHigh,
		},
		"accepts_iterations_within_policy": {
			iterations:     150,
			keyTagToDNSKey: dnsKeysByTag{12345: {ed25519Key}},
			expectLen:      1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			nsec3RR := &dns.NSEC3{
				Hdr:        dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeNSEC3, Class: dns.ClassINET},
				Hash:       dns.SHA1,
				Flags:      0,
				Iterations: tc.iterations,
				Salt:       "ABCD",
			}

			sanitized, err := nsec3InitialChecks([]dns.RR{nsec3RR}, tc.keyTagToDNSKey)

			if tc.errWrapped != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.errWrapped)
			} else {
				require.NoError(t, err)
				assert.Len(t, sanitized, tc.expectLen)
			}
		})
	}
}
