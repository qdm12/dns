package dnssec

import (
	"fmt"

	"github.com/miekg/dns"
)

func mustRRToDNSKey(rr dns.RR) *dns.DNSKEY {
	dnsKey, ok := rr.(*dns.DNSKEY)
	if !ok {
		panic(fmt.Sprintf("RR is of type %T and not of type *dns.DNSKEY", rr))
	}
	return dnsKey
}

// makeKeyTagToDNSKey creates a map of key tag to DNSKEY from a DNSKEY RRSet,
// ignoring any RR which is not a Zone signing key.
func makeKeyTagToDNSKey(dnsKeyRRSet []dns.RR) (keyTagToDNSKey map[uint16]*dns.DNSKEY) {
	keyTagToDNSKey = make(map[uint16]*dns.DNSKEY, len(dnsKeyRRSet))
	for _, dnsKeyRR := range dnsKeyRRSet {
		dnsKey := mustRRToDNSKey(dnsKeyRR)
		if dnsKey.Flags&dns.ZONE == 0 {
			// As described in https://datatracker.ietf.org/doc/html/rfc4034#section-2.1.1
			// and https://datatracker.ietf.org/doc/html/rfc4034#section-5.2:
			// If bit 7 has value 0, then the DNSKEY record holds some other type of DNS
			// public key and MUST NOT be used to verify RRSIGs that cover RRsets.
			// The DNSKEY RR Flags MUST have Flags bit 7 set. If the
			// DNSKEY flags do not indicate a DNSSEC zone key, the DS
			// RR (and the DNSKEY RR it references) MUST NOT be used
			// in the validation process.
			continue
		}
		keyTagToDNSKey[dnsKey.KeyTag()] = dnsKey
	}
	return keyTagToDNSKey
}

const (
	algoPreferenceRecommended uint8 = iota
	algoPreferenceMust
	algoPreferenceMay
	algoPreferenceMustNot
	algoPreferenceUnknown
)

// lessDNSKeyAlgorithm returns true if algoID1 < algoID2 in terms
// of preference. The preference is determined by the table defined in:
// https://datatracker.ietf.org/doc/html/rfc8624#section-3.1
func lessDNSKeyAlgorithm(algoID1, algoID2 uint8) bool {
	return algoIDToPreference(algoID1) < algoIDToPreference(algoID2)
}

// algoIDToPreference returns the preference level of the algorithm ID.
// Note this is a function with a switch statement, which not only provide
// immutability compared to a global variable map, but is also x10 faster
// than map lookups.
func algoIDToPreference(algoID uint8) (preference uint8) {
	switch algoID {
	case dns.RSAMD5, dns.DSA, dns.DSANSEC3SHA1:
		return algoPreferenceMustNot
	case dns.ECCGOST:
		return algoPreferenceMay
	case dns.RSASHA1, dns.RSASHA1NSEC3SHA1, dns.RSASHA256, dns.RSASHA512, dns.ECDSAP256SHA256:
		return algoPreferenceMust
	case dns.ECDSAP384SHA384, dns.ED25519, dns.ED448:
		return algoPreferenceRecommended
	default:
		return algoPreferenceUnknown
	}
}
