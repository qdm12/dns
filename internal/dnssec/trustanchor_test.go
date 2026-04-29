package dnssec

import (
	"strings"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rootKSK2017PublicKey() string {
	parts := []string{
		"AwEAAaz/tAm8yTn4Mfeh5eyI96WSVexTBAvkMgJzkKTOiW1vkIbzxeF3+/4RgWOq7",
		"HrxRixHlFlExOLAJr5emLvN7SWXgnLh4+B5xQlNVz8Og8kvArMtNROxVQuCaSnIDdD5LKyWbRd2n",
		"9WGe2R8PzgCmr3EgVLrjyBxWezF0jLHwVN8efS3rCj/EWgvIWgb9tarpVUDK/b58Da+sqqls3eNbu",
		"v7pr+eoZG+SrDK6nWeL3c6H5Apxz7LjVc1uTIdsIXxuOLYA4/ilBmSVIzuDWfdRUfhHdY6+cn8HFR",
		"m+2hM8AnXGXws9555KrUB5qihylGa8subX2Nn6UwNR1AkUTV74bU=",
	}
	return strings.Join(parts, "")
}

func Test_defaultRootTrustAnchors(t *testing.T) {
	t.Parallel()

	anchors := defaultRootTrustAnchors()
	require.Len(t, anchors, 2)
	assert.Equal(t, rootKSK2017KeyTag, anchors[0].KeyTag)
	assert.Equal(t, rootKSK2024KeyTag, anchors[1].KeyTag)
}

func Test_verifyRootTrustAnchors(t *testing.T) {
	t.Parallel()

	dnsKey := &dns.DNSKEY{
		Hdr:       dns.RR_Header{Name: ".", Rrtype: dns.TypeDNSKEY, Class: dns.ClassINET},
		Flags:     257,
		Protocol:  3,
		Algorithm: dns.RSASHA256,
		PublicKey: rootKSK2017PublicKey(),
	}
	validAnchor := *dnsKey.ToDS(dns.SHA256)
	invalidAnchor := validAnchor
	invalidAnchor.KeyTag++

	keyTagToDNSKeys := dnsKeysByTag{dnsKey.KeyTag(): {dnsKey}}

	err := verifyRootTrustAnchors([]dns.DS{invalidAnchor, validAnchor}, keyTagToDNSKeys)
	assert.NoError(t, err)

	err = verifyRootTrustAnchors([]dns.DS{invalidAnchor}, keyTagToDNSKeys)
	require.Error(t, err)
	assert.ErrorContains(t, err, "for root anchor key tag")
	assert.ErrorContains(t, err, errDNSKeyNotFound.Error())
}

func Test_deriveTrustAnchorsFromDNSKEYRRSet(t *testing.T) {
	t.Parallel()

	ksk := &dns.DNSKEY{
		Hdr:       dns.RR_Header{Name: ".", Rrtype: dns.TypeDNSKEY, Class: dns.ClassINET},
		Flags:     257,
		Protocol:  3,
		Algorithm: dns.RSASHA256,
		PublicKey: rootKSK2017PublicKey(),
	}
	zsk := &dns.DNSKEY{
		Hdr:       dns.RR_Header{Name: ".", Rrtype: dns.TypeDNSKEY, Class: dns.ClassINET},
		Flags:     256,
		Protocol:  3,
		Algorithm: dns.RSASHA256,
		PublicKey: ksk.PublicKey,
	}

	dsRecords := deriveTrustAnchorsFromDNSKEYRRSet([]dns.RR{ksk, zsk, ksk})
	require.Len(t, dsRecords, 1)
	assert.Equal(t, ksk.KeyTag(), dsRecords[0].KeyTag)
}
