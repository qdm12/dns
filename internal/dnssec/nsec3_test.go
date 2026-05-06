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
