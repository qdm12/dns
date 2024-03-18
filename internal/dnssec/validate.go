package dnssec

import (
	"errors"
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

// verify uses the zone data in the signed zone and its parent signed zones
// to verify the DNSSEC chain of trust.
// It starts the verification with the RRSet given as argument, and,
// assuming a signature is valid, it walks through the slice of signed
// zones checking the RRSIGs on the DNSKEY and DS resource record sets.
func validateWithChain(desiredZone string, qType uint16,
	desiredResponse dnssecResponse, chain []signedData) (err error) {
	// Verify the root zone "."
	rootZone := chain[0]

	// Verify DNSKEY RRSet with its RRSIG and the DNSKEY matching
	// the RRSIG key tag.
	rootZoneKeyTagToDNSKey := makeKeyTagToDNSKey(rootZone.dnsKeyResponse.onlyAnswerRRSet())
	err = verifyRRSetRRSigs(rootZone.dnsKeyResponse.onlyAnswerRRSet(),
		rootZone.dnsKeyResponse.onlyAnswerRRSigs(), rootZoneKeyTagToDNSKey)
	if err != nil {
		return fmt.Errorf("verifying DNSKEY records for the root zone: %w",
			err)
	}

	// Verify the root anchor digest against the digest of the DS
	// calculated from the DNSKEY of the root zone matching the
	// root anchor key tag.
	const (
		rootAnchorKeyTag = 20326
		rootAnchorDigest = "E06D44B80B8F1D39A95C0B0D7C65D08458E880409BBC683457104237C7F8EC8D"
	)
	rootAnchor := &dns.DS{
		Algorithm:  dns.RSASHA256,
		DigestType: dns.SHA256,
		KeyTag:     rootAnchorKeyTag,
		Digest:     rootAnchorDigest,
	}
	err = verifyDS(rootAnchor, rootZoneKeyTagToDNSKey)
	if err != nil {
		return fmt.Errorf("verifying the root anchor: %w", err)
	}

	parentZoneInsecure := false
	for i := 1; i < len(chain); i++ {
		// Iterate in this order: "com.", "example.com.", "abc.example.com."
		// Note the chain may not include the desired zone if one of its parent
		// zone is unsigned. Checking a parent zone is indeed unsigned
		// with DS-associated NSEC/NSEC3 RRSets also verifies the desired
		// zone is unsigned.
		zoneData := chain[i]
		parentZoneData := chain[i-1]

		switch {
		case zoneData.dsResponse.isNXDomain():
			parentKeyTagToDNSKey := makeKeyTagToDNSKey(parentZoneData.dnsKeyResponse.onlyAnswerRRSet())
			err = validateNxDomain(zoneData.zone, zoneData.dsResponse.authorityRRSets, parentKeyTagToDNSKey)
			if err != nil {
				return fmt.Errorf("validating NXDOMAIN DS response: %w", err)
			}
			// no need to continue the verification for this zone since
			// child zones are unsigned.
			parentZoneInsecure = true
		case zoneData.dsResponse.isNoData():
			parentKeyTagToDNSKey := makeKeyTagToDNSKey(parentZoneData.dnsKeyResponse.onlyAnswerRRSet())
			err = validateNoDataDS(zoneData.zone, zoneData.dsResponse.authorityRRSets, parentKeyTagToDNSKey)
			if err != nil {
				return fmt.Errorf("validating no data DS response: %w", err)
			}

			// no need to continue the verification for this zone since
			// child zones are unsigned.
			parentZoneInsecure = true
		default: // signed zone
		}

		if parentZoneInsecure {
			break
		}

		// Validate DNSKEY RRSet with its RRSIG and the DNSKEY matching
		// the RRSIG key tag. Note a zone should only have a DNSKEY RRSet
		// if it has a DS RRSet.
		keyTagToDNSKey := makeKeyTagToDNSKey(zoneData.dnsKeyResponse.onlyAnswerRRSet())
		err = verifyRRSetRRSigs(zoneData.dnsKeyResponse.onlyAnswerRRSet(),
			zoneData.dnsKeyResponse.onlyAnswerRRSigs(),
			keyTagToDNSKey)
		if err != nil {
			return fmt.Errorf("validating DNSKEY RRSet for zone %s: %w",
				zoneData.zone, err)
		}

		// Validate DS RRSet with its RRSIG and the DNSKEY of its parent zone
		// matching the RRSIG key tag.
		parentKeyTagToDNSKey := makeKeyTagToDNSKey(parentZoneData.dnsKeyResponse.onlyAnswerRRSet())
		err = verifyRRSetRRSigs(zoneData.dsResponse.onlyAnswerRRSet(),
			zoneData.dsResponse.onlyAnswerRRSigs(), parentKeyTagToDNSKey)
		if err != nil {
			return fmt.Errorf("validating DS RRSet for zone %s: %w",
				zoneData.zone, err)
		}

		// Validate DS RRSet digests with their corresponding DNSKEYs.
		err = verifyDSRRSet(zoneData.dsResponse.onlyAnswerRRSet(), keyTagToDNSKey)
		if err != nil {
			return fmt.Errorf("verifying DS RRSet for zone %s: %w",
				zoneData.zone, err)
		}
	}

	if !desiredResponse.isSigned() && !parentZoneInsecure {
		// The desired query returned an insecure response
		// (unsigned answers or no NSEC/NSEC3 RRSets) and
		// no parent zone was found to be unsigned, meaning this
		// is bogus.
		return fmt.Errorf("%w: desired query response is unsigned "+
			"but no parent zone was found to be insecure", ErrBogus)
	}

	if parentZoneInsecure {
		// Whether the desired query is signed or not, if a parent zone
		// is insecure, the desired query is insecure.
		// For example IN A textsecure-service.whispersystems.org. has NSEC
		// signed by whispersystems.org., which has DNSKEYs but no DS record.
		return nil
	}

	// From this point, the desiredResponse is signed.

	// Note we validate the desired zone last since there might be a
	// break in the chain, where there is no DNSKEY for the parent zone
	// of the desired zone which has a DS RRSet.
	// For example for textsecure-service.whispersystems.org.
	var lastSecureZoneData signedData
	for i := len(chain) - 1; i >= 0; i-- {
		zoneData := chain[i]
		if len(zoneData.dsResponse.onlyAnswerRRSet()) > 0 {
			lastSecureZoneData = zoneData
			break
		}
	}

	lastSecureKeyTagToDNSKey := makeKeyTagToDNSKey(lastSecureZoneData.dnsKeyResponse.onlyAnswerRRSet())
	switch {
	case desiredResponse.rcode == dns.RcodeNameError: // NXDOMAIN
		err = validateNxDomain(desiredZone, desiredResponse.authorityRRSets,
			lastSecureKeyTagToDNSKey)
		if err != nil {
			return fmt.Errorf("validating negative NXDOMAIN response: %w", err)
		}
	case len(desiredResponse.answerRRSets) == 0: // NODATA
		err = validateNoData(desiredZone, qType, desiredResponse.authorityRRSets,
			lastSecureKeyTagToDNSKey)
		if err != nil {
			return fmt.Errorf("validating negative NODATA response: %w", err)
		}
	default:
		// Verify the desired RRSets with the DNSKEY of the desired
		// zone matching the RRSIG key tag.
		err = verifyRRSetsRRSig(desiredResponse.answerRRSets,
			desiredResponse.authorityRRSets, lastSecureKeyTagToDNSKey)
		if err != nil {
			return fmt.Errorf("verifying RRSets with RRSigs: %w", err)
		}
	}

	return nil
}

// verifyDSRRSet verifies the digest of each received DS
// is equal to the digest of the calculated DS obtained
// from the DNSKEY (KSK) matching the received DS key tag.
func verifyDSRRSet(dsRRSet []dns.RR,
	keyTagToDNSKey map[uint16]*dns.DNSKEY) (err error) {
	for _, rr := range dsRRSet {
		ds := mustRRToDS(rr)
		err = verifyDS(ds, keyTagToDNSKey)
		if err != nil {
			return fmt.Errorf("verifying DS record: %w", err)
		}
	}
	return nil
}

var (
	ErrDNSKeyNotFound   = errors.New("DNSKEY resource record not found")
	ErrDNSKeyToDS       = errors.New("failed to calculate DS from DNSKEY")
	ErrDNSKeyDSMismatch = errors.New("DS does not match DNS key")
)

func verifyDS(receivedDS *dns.DS,
	keyTagToDNSKey map[uint16]*dns.DNSKEY) error {
	// Note keyTagToDNSKey only contains ZSKs.
	dnsKey, ok := keyTagToDNSKey[receivedDS.KeyTag]
	if !ok {
		return fmt.Errorf("for RRSIG key tag %d: %w",
			receivedDS.KeyTag, ErrDNSKeyNotFound)
	}

	calculatedDS := dnsKey.ToDS(receivedDS.DigestType)
	if calculatedDS == nil {
		return fmt.Errorf("%w: for DNSKEY name %s and digest type %d",
			ErrDNSKeyToDS, dnsKey.Header().Name, receivedDS.DigestType)
	}

	if !strings.EqualFold(receivedDS.Digest, calculatedDS.Digest) {
		return fmt.Errorf("%w: DS record has digest %s "+
			"but DNSKEY calculated DS has digest %s",
			ErrDNSKeyDSMismatch, receivedDS.Digest, calculatedDS.Digest)
	}

	return nil
}
