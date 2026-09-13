package dnssec

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_verifyDSRRSet(t *testing.T) {
	t.Parallel()

	dnsKey := &dns.DNSKEY{
		Hdr: dns.RR_Header{
			Name:   "in-addr.arpa.",
			Rrtype: dns.TypeDNSKEY,
			Class:  dns.ClassINET,
		},
		Flags:     dns.ZONE,
		Protocol:  3,
		Algorithm: dns.ED25519,
		PublicKey: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
	}

	validDS := dnsKey.ToDS(dns.SHA256)
	require.NotNil(t, validDS)
	validDS.Hdr = dns.RR_Header{
		Name:   "in-addr.arpa.",
		Rrtype: dns.TypeDS,
		Class:  dns.ClassINET,
	}

	staleDS := &dns.DS{
		Hdr: dns.RR_Header{
			Name:   "in-addr.arpa.",
			Rrtype: dns.TypeDS,
			Class:  dns.ClassINET,
		},
		KeyTag:     validDS.KeyTag + 1,
		Algorithm:  validDS.Algorithm,
		DigestType: validDS.DigestType,
		Digest:     validDS.Digest,
	}

	keyTagToDNSKeys := dnsKeysByTag{
		dnsKey.KeyTag(): []*dns.DNSKEY{dnsKey},
	}

	t.Run("accepts_if_one_ds_matches", func(t *testing.T) {
		t.Parallel()

		err := verifyDSRRSet([]dns.RR{staleDS, validDS}, keyTagToDNSKeys)

		require.NoError(t, err)
	})

	t.Run("fails_if_no_ds_matches", func(t *testing.T) {
		t.Parallel()

		err := verifyDSRRSet([]dns.RR{staleDS}, keyTagToDNSKeys)

		require.Error(t, err)
		assert.ErrorContains(t, err, "no DS record matched child DNSKEY RRSet")
		assert.ErrorIs(t, err, errDNSKeyNotFound)
	})
}
