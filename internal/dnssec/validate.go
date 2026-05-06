package dnssec

import (
	"errors"
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

var errWildcardedDNAME = errors.New("DNAME record cannot be synthesized from wildcard per RFC 6672 section 5.3.1")

// verify uses the zone data in the signed zone and its parent signed zones
// to verify the DNSSEC chain of trust.
// It starts the verification with the RRSet given as argument, and,
// assuming a signature is valid, it walks through the slice of signed
// zones checking the RRSIGs on the DNSKEY and DS resource record sets.
func validateWithChain(desiredZone string, qType uint16,
	desiredResponse dnssecResponse, chain []signedData, rootTrustAnchors []dns.DS,
) (err error) {
	err = validateRootZone(chain[0], rootTrustAnchors)
	if err != nil {
		return fmt.Errorf("validating root zone: %w", err)
	}

	wildcardName, chain, err := applyWildcardChainRestriction(desiredResponse, chain)
	if err != nil {
		return fmt.Errorf("applying wildcard chain restriction: %w", err)
	}

	parentZoneInsecure, err := validateParentChain(chain)
	if err != nil {
		return fmt.Errorf("validating parent chain: %w", err)
	}

	if !desiredResponse.isSigned() && !parentZoneInsecure {
		// The desired query returned an insecure response
		// (unsigned answers or no NSEC/NSEC3 RRSets) and
		// no parent zone was found to be unsigned, meaning this
		// is bogus.
		return fmt.Errorf("%w: desired query response is unsigned "+
			"but no parent zone was found to be insecure", errBogus)
	}

	if parentZoneInsecure {
		// Whether the desired query is signed or not, if a parent zone
		// is insecure, the desired query is insecure.
		// For example IN A textsecure-service.whispersystems.org. has NSEC
		// signed by whispersystems.org., which has DNSKEYs but no DS record.
		return nil
	}

	// From this point, the desiredResponse is signed.

	lastSecureZoneData := findLastSecureZoneData(chain)
	lastSecureKeyTagToDNSKeys := makeKeyTagToDNSKeys(lastSecureZoneData.dnsKeyResponse.onlyAnswerRRSet())

	err = validateDesiredResponse(desiredZone, qType, desiredResponse,
		wildcardName, lastSecureKeyTagToDNSKeys)
	if err != nil {
		return fmt.Errorf("validating desired response: %w", err)
	}

	return nil
}

func validateRootZone(rootZone signedData,
	rootTrustAnchors []dns.DS,
) (err error) {
	// Verify DNSKEY RRSet with its RRSIG and the DNSKEY matching
	// the RRSIG key tag.
	rootZoneKeyTagToDNSKeys := makeKeyTagToDNSKeys(rootZone.dnsKeyResponse.onlyAnswerRRSet())
	err = verifyRRSetRRSigs(rootZone.dnsKeyResponse.onlyAnswerRRSet(),
		rootZone.dnsKeyResponse.onlyAnswerRRSigs(), rootZoneKeyTagToDNSKeys,
		newRRSIGValidationBudget())
	if err != nil {
		return fmt.Errorf("verifying DNSKEY records for the root zone: %w", err)
	}

	// Verify that the configured root trust anchors match at least one KSK
	// published in the authenticated root DNSKEY RRSet.
	err = verifyRootTrustAnchors(rootTrustAnchors, rootZoneKeyTagToDNSKeys)
	if err != nil {
		return fmt.Errorf("verifying the root trust anchors: %w", err)
	}

	return nil
}

func applyWildcardChainRestriction(desiredResponse dnssecResponse,
	chain []signedData,
) (wildcardName string, restrictedChain []signedData, err error) {
	wildcardName = extractWildcardExpansion(desiredResponse.answerRRSets)
	if wildcardName == "" {
		return "", chain, nil
	}

	wildcardLabelsCount := dns.CountLabel(wildcardName)
	restrictedChain = chain[:wildcardLabelsCount]

	// Per RFC 6672 section 5.3.1, DNAME records must not be synthesized
	// from wildcards. Reject any response with wildcard expansion that
	// contains DNAME records in the answer section.
	if answerHasWildcardedDNAME(desiredResponse.answerRRSets) {
		return "", nil, fmt.Errorf("%w: answer contains DNAME with wildcard %s",
			errWildcardedDNAME, wildcardName)
	}

	return wildcardName, restrictedChain, nil
}

func validateParentChain(chain []signedData) (parentZoneInsecure bool, err error) {
	for i := 1; i < len(chain); i++ {
		// Iterate in this order: "com.", "example.com.", "abc.example.com."
		// Note the chain may not include the desired zone if one of its parent
		// zone is unsigned. Checking a parent zone is indeed unsigned
		// with DS-associated NSEC/NSEC3 RRSets also verifies the desired
		// zone is unsigned.
		zoneData := chain[i]
		parentZoneData := chain[i-1]

		parentZoneInsecure, err = validateDSResponse(zoneData, parentZoneData)
		if err != nil {
			return false, err
		}
		if parentZoneInsecure {
			return true, nil
		}

		err = validateSignedDelegation(zoneData, parentZoneData)
		if err != nil {
			if errors.Is(err, errRRSigAllBadKeys) {
				// Some delegated zones publish DNSKEY RRSIGs that cannot be
				// validated by available keys. Treat this zone as insecure
				// rather than failing the entire response validation.
				return true, nil
			}
			return false, err
		}
	}

	return false, nil
}

func validateDSResponse(zoneData, parentZoneData signedData,
) (parentZoneInsecure bool, err error) {
	switch {
	case zoneData.dsResponse.isNXDomain():
		if !zoneData.dsResponse.isSigned() {
			// Some resolvers may return unsigned NXDOMAIN for DS queries
			// without NSEC/NSEC3 proofs. Treat this as insecure delegation.
			return true, nil
		}

		parentKeyTagToDNSKeys := makeKeyTagToDNSKeys(parentZoneData.dnsKeyResponse.onlyAnswerRRSet())
		err = validateNxDomain(zoneData.zone, zoneData.dsResponse.authorityRRSets, parentKeyTagToDNSKeys)
		if err != nil {
			return false, fmt.Errorf("validating NXDOMAIN DS response: %w", err)
		}

		// No need to continue the verification for this zone since child zones
		// are unsigned.
		return true, nil
	case zoneData.dsResponse.isNoData():
		if !zoneData.dsResponse.isSigned() {
			// Some resolvers may return unsigned NODATA for DS queries
			// without NSEC/NSEC3 proofs. Treat this as insecure delegation.
			return true, nil
		}

		parentKeyTagToDNSKeys := makeKeyTagToDNSKeys(parentZoneData.dnsKeyResponse.onlyAnswerRRSet())
		err = validateNoDataDS(zoneData.zone, zoneData.dsResponse.authorityRRSets, parentKeyTagToDNSKeys)
		if err != nil {
			return false, fmt.Errorf("validating no data DS response: %w", err)
		}

		// No need to continue the verification for this zone since child zones
		// are unsigned.
		return true, nil
	case !zoneData.dsResponse.isSigned():
		// The DS query returned an unsigned answer that is neither NODATA nor
		// NXDOMAIN (e.g. a CNAME chain for a name that is an alias, not a
		// delegation). Treat this as an insecure delegation.
		return true, nil
	default:
		return false, nil
	}
}

func validateSignedDelegation(zoneData, parentZoneData signedData) (err error) {
	// Validate DNSKEY RRSet with its RRSIG and the DNSKEY matching
	// the RRSIG key tag. Note a zone should only have a DNSKEY RRSet
	// if it has a DS RRSet.
	keyTagToDNSKeys := makeKeyTagToDNSKeys(zoneData.dnsKeyResponse.onlyAnswerRRSet())
	err = verifyRRSetRRSigs(zoneData.dnsKeyResponse.onlyAnswerRRSet(),
		zoneData.dnsKeyResponse.onlyAnswerRRSigs(),
		keyTagToDNSKeys, newRRSIGValidationBudget())
	if err != nil {
		return fmt.Errorf("validating DNSKEY RRSet for zone %s: %w", zoneData.zone, err)
	}

	// Validate DS RRSet with its RRSIG and the DNSKEY of its parent zone
	// matching the RRSIG key tag.
	parentKeyTagToDNSKeys := makeKeyTagToDNSKeys(parentZoneData.dnsKeyResponse.onlyAnswerRRSet())
	err = verifyRRSetRRSigs(zoneData.dsResponse.onlyAnswerRRSet(),
		zoneData.dsResponse.onlyAnswerRRSigs(), parentKeyTagToDNSKeys,
		newRRSIGValidationBudget())
	if err != nil {
		return fmt.Errorf("validating DS RRSet for zone %s: %w", zoneData.zone, err)
	}

	// Validate DS RRSet digests with their corresponding DNSKEYs.
	err = verifyDSRRSet(zoneData.dsResponse.onlyAnswerRRSet(), keyTagToDNSKeys)
	if err != nil {
		return fmt.Errorf("verifying DS RRSet for zone %s: %w", zoneData.zone, err)
	}

	return nil
}

func findLastSecureZoneData(chain []signedData) (lastSecureZoneData signedData) {
	// Note we validate the desired zone last since there might be a
	// break in the chain, where there is no DNSKEY for the parent zone
	// of the desired zone which has a DS RRSet.
	// For example for textsecure-service.whispersystems.org.
	for i := len(chain) - 1; i >= 0; i-- {
		zoneData := chain[i]
		if i == 0 || len(zoneData.dsResponse.answerRRSets) == 1 {
			return zoneData
		}
	}

	return signedData{}
}

func validateDesiredResponse(desiredZone string, qType uint16,
	desiredResponse dnssecResponse, wildcardName string,
	lastSecureKeyTagToDNSKeys dnsKeysByTag,
) (err error) {
	switch {
	case desiredResponse.isNXDomain():
		err = validateNxDomain(desiredZone, desiredResponse.authorityRRSets,
			lastSecureKeyTagToDNSKeys)
		if err != nil {
			return fmt.Errorf("validating negative NXDOMAIN response: %w", err)
		}
		return nil
	case desiredResponse.isNoData():
		err = validateNoData(desiredZone, qType, desiredResponse.authorityRRSets,
			lastSecureKeyTagToDNSKeys)
		if err != nil {
			return fmt.Errorf("validating negative NODATA response: %w", err)
		}
		return nil
	default:
		// Verify the desired RRSets with the DNSKEY of the desired
		// zone matching the RRSIG key tag.
		err = verifyRRSetsRRSig(desiredResponse.answerRRSets, lastSecureKeyTagToDNSKeys)
		if err != nil {
			return fmt.Errorf("verifying RRSets with RRSigs: %w", err)
		}

		if wildcardName == "" { // no wildcard expansion
			return nil
		}

		err = verifyRRSetsRRSig(desiredResponse.authorityRRSets, lastSecureKeyTagToDNSKeys)
		if err != nil {
			return fmt.Errorf("verifying authority section RRSets with RRSigs: %w", err)
		}

		err = validateWildcardExpansion(desiredZone, desiredResponse.authorityRRSets,
			lastSecureKeyTagToDNSKeys)
		if err != nil {
			return fmt.Errorf("validating wildcard expansion: %w", err)
		}

		return nil
	}
}

var errNoDSRecordToVerify = errors.New("no DS record to verify")

// verifyDSRRSet verifies the DS RRSet against child DNSKEYs.
//
// The RRSet is accepted if at least one DS record matches a DNSKEY
// in the child zone. This tolerates stale DS records during rollover,
// as long as there is a valid DS->DNSKEY path.
func verifyDSRRSet(dsRRSet []dns.RR,
	keyTagToDNSKeys dnsKeysByTag,
) (err error) {
	errs := new(joinedErrors)
	var oneMatched bool

	for _, rr := range dsRRSet {
		ds := mustRRToDS(rr)
		err = verifyDS(ds, keyTagToDNSKeys)
		if err != nil {
			errs.add(fmt.Errorf("DS key tag %d: %w", ds.KeyTag, err))
			continue
		}

		oneMatched = true
	}

	if oneMatched {
		return nil
	}

	if len(errs.errs) == 0 {
		return fmt.Errorf("%w", errNoDSRecordToVerify)
	}

	return fmt.Errorf("no DS record matched child DNSKEY RRSet: %w", errs)
}

var (
	errDNSKeyNotFound   = errors.New("DNSKEY resource record not found")
	errDNSKeyDSMismatch = errors.New("DS does not match DNS key")
)

func verifyDS(receivedDS *dns.DS,
	keyTagToDNSKeys dnsKeysByTag,
) error {
	dnsKeys := keyTagToDNSKeys[receivedDS.KeyTag]
	if len(dnsKeys) == 0 {
		return fmt.Errorf("for DS key tag %d: %w",
			receivedDS.KeyTag, errDNSKeyNotFound)
	}

	for _, dnsKey := range dnsKeys {
		if dnsKey.Algorithm != receivedDS.Algorithm {
			continue
		}

		calculatedDS := dnsKey.ToDS(receivedDS.DigestType)
		if calculatedDS == nil {
			continue
		}

		if strings.EqualFold(receivedDS.Digest, calculatedDS.Digest) {
			return nil
		}
	}

	return fmt.Errorf("%w: for key tag %d and algorithm %d",
		errDNSKeyDSMismatch, receivedDS.KeyTag, receivedDS.Algorithm)
}
