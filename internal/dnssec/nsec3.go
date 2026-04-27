package dnssec

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/miekg/dns"
	"golang.org/x/exp/maps"
)

func mustRRToNSEC3(rr dns.RR) (nsec3 *dns.NSEC3) {
	nsec3, ok := rr.(*dns.NSEC3)
	if !ok {
		panic(fmt.Sprintf("RR is of type %T and not of type *dns.NSEC3", rr))
	}
	return nsec3
}

// extractNSEC3s returns the NSEC3 RRs found in the NSEC3
// signed RRSet from the slice of signed RRSets. It also returns
// wildcard as true if one of the NSEC3 RRSets RRSigs is for a wildcard.
func extractNSEC3s(rrSets []dnssecRRSet) (
	rrs []dns.RR, wildcard bool,
) {
	rrs = make([]dns.RR, 0, len(rrSets))
	for _, rrSet := range rrSets {
		if rrSet.qtype() != dns.TypeNSEC3 {
			continue
		}
		rrs = append(rrs, rrSet.rrSet...)

		if !wildcard {
			wildcard = slices.ContainsFunc(rrSet.rrSigs, isRRSigForWildcard)
		}
	}
	return rrs, wildcard
}

var (
	errNSEC3RRSetDifferentHashTypes  = errors.New("NSEC3 RRSet contains different hash types")
	errNSEC3RRSetDifferentIterations = errors.New("NSEC3 RRSet contains different iterations")
	errNSEC3RRSetDifferentSalts      = errors.New("NSEC3 RRSet contains different salts")
	errNSEC3IterationsTooHigh        = errors.New("NSEC3 iteration count too high for DNSKEY strength policy")
)

func nsec3InitialChecks(nsec3RRSet []dns.RR, keyTagToDNSKey map[uint16]*dns.DNSKEY,
) (sanitizedNSEC3RRSet []dns.RR, err error) {
	sanitizedNSEC3RRSet = make([]dns.RR, 0, len(nsec3RRSet))
	maxIterations := maxNSEC3IterationsForDNSKeys(keyTagToDNSKey)

	const usualCapacity = 1
	hashTypes := make(map[uint8]struct{}, usualCapacity)
	iterations := make(map[uint16]struct{}, usualCapacity)
	salts := make(map[string]struct{}, usualCapacity)

	for _, nsec3RR := range nsec3RRSet {
		nsec3 := mustRRToNSEC3(nsec3RR)

		// Only accept supported hash type
		// https://datatracker.ietf.org/doc/html/rfc5155#section-8.1
		if nsec3.Hash != dns.SHA1 {
			continue
		}

		// Flag field must be zero or one (opt-out).
		// https://datatracker.ietf.org/doc/html/rfc5155#section-8.2
		if nsec3.Flags != 0 && nsec3.Flags != 1 {
			continue
		}

		if nsec3.Iterations > maxIterations {
			return nil, fmt.Errorf("%w: got %d and policy max is %d",
				errNSEC3IterationsTooHigh, nsec3.Iterations, maxIterations)
		}

		// Track hash algorithms, iterations and salts
		// https://datatracker.ietf.org/doc/html/rfc5155#section-8.2
		hashTypes[nsec3.Hash] = struct{}{}
		iterations[nsec3.Iterations] = struct{}{}
		salts[nsec3.Salt] = struct{}{}

		sanitizedNSEC3RRSet = append(sanitizedNSEC3RRSet, nsec3RR)
	}

	// Verify all NSEC3 RRSet RRs have the same hash type, iterations and salt
	// If not, the response may be considered as bogus, so we return an error.
	// https://datatracker.ietf.org/doc/html/rfc5155#section-8.2
	switch {
	case len(hashTypes) > 1:
		return nil, fmt.Errorf("%w: %s", errNSEC3RRSetDifferentHashTypes,
			hashesToString(maps.Keys(hashTypes)))
	case len(iterations) > 1:
		return nil, fmt.Errorf("%w: %s", errNSEC3RRSetDifferentIterations,
			integersToString(maps.Keys(iterations)))
	case len(salts) > 1:
		return nil, fmt.Errorf("%w: %s", errNSEC3RRSetDifferentSalts,
			strings.Join(maps.Keys(salts), ", "))
	}

	return sanitizedNSEC3RRSet, nil
}

func maxNSEC3IterationsForDNSKeys(keyTagToDNSKey map[uint16]*dns.DNSKEY) uint16 {
	// RFC 5155 section 10.3 policy table constants.
	const (
		nsec3SmallKeyBitsThreshold  uint16 = 1024
		nsec3MediumKeyBitsThreshold uint16 = 2048
		nsec3MaxIterationsSmallKey  uint16 = 150
		nsec3MaxIterationsMediumKey uint16 = 500
		nsec3MaxIterationsLargeKey  uint16 = 2500
	)

	if len(keyTagToDNSKey) == 0 {
		return nsec3MaxIterationsLargeKey
	}

	var minBits uint16
	for _, dnsKey := range keyTagToDNSKey {
		bits := dnsKeyStrengthBits(dnsKey)
		if bits == 0 {
			continue
		}
		if minBits == 0 || bits < minBits {
			minBits = bits
		}
	}

	if minBits == 0 {
		return nsec3MaxIterationsLargeKey
	}

	// RFC 5155 section 10.3 policy table.
	if minBits <= nsec3SmallKeyBitsThreshold {
		return nsec3MaxIterationsSmallKey
	}
	if minBits <= nsec3MediumKeyBitsThreshold {
		return nsec3MaxIterationsMediumKey
	}
	return nsec3MaxIterationsLargeKey
}

//nolint:mnd
func dnsKeyStrengthBits(dnsKey *dns.DNSKEY) uint16 {
	if dnsKey == nil {
		return 0
	}

	switch dnsKey.Algorithm {
	case dns.ECDSAP256SHA256, dns.ED25519:
		return 256
	case dns.ECDSAP384SHA384:
		return 384
	case dns.ED448:
		return 456
	case dns.RSASHA1, dns.RSASHA1NSEC3SHA1, dns.RSASHA256, dns.RSASHA512:
		return rsaDNSKEYBits(dnsKey.PublicKey)
	default:
		return 0
	}
}

func rsaDNSKEYBits(publicKeyBase64 string) uint16 {
	publicKey, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	const minPublicKeyLength = 3
	if err != nil || len(publicKey) < minPublicKeyLength {
		return 0
	}

	var exponentLength int
	var offset int
	if publicKey[0] == 0 {
		exponentLength = int(binary.BigEndian.Uint16(publicKey[1:3]))
		offset = 3
	} else {
		exponentLength = int(publicKey[0])
		offset = 1
	}

	modulusOffset := offset + exponentLength
	if exponentLength <= 0 || modulusOffset >= len(publicKey) {
		return 0
	}

	modulusLength := len(publicKey) - modulusOffset
	if modulusLength <= 0 {
		return 0
	}

	return uint16(modulusLength * 8) //nolint:gosec,mnd
}

func nsec3ValidateNxDomain(qname string, nsec3RRSet []dns.RR) (err error) {
	// Proof qname does not exist with the closest encloser proof
	closestEncloser, err := nsec3VerifyClosestEncloserProof(qname, nsec3RRSet)
	if err != nil {
		return fmt.Errorf("for qname %s: "+
			"validating closest encloser proof: %w",
			qname, err)
	}

	// Proof the wildcard matching qname does not exist
	wildcardName := "*." + closestEncloser
	wildcardCoveringNSEC3 := nsec3FindCovering(wildcardName, nsec3RRSet)
	if wildcardCoveringNSEC3 == nil {
		return fmt.Errorf("for qname %s: %w: "+
			"NSEC3 matching wildcard %s not found",
			qname, errBogus, wildcardName)
	}

	return nil
}

// nsec3ValidateNoData validates a no data response for a given QTYPE.
// See https://datatracker.ietf.org/doc/html/rfc5155#section-8.5
// and https://datatracker.ietf.org/doc/html/rfc5155#section-8.6
func nsec3ValidateNoData(qname string, qType uint16,
	nsec3RRSet []dns.RR,
) (err error) {
	if qType == dns.TypeDS {
		return nsec3ValidateNoDataDS(qname, nsec3RRSet)
	}

	err = nsec3RRSetHasMatchingWithoutTypes(nsec3RRSet,
		qname, qType, dns.TypeCNAME)
	if err != nil {
		return fmt.Errorf("for qname %s: %w", qname, err)
	}
	return nil
}

// nsec3ValidateNoDataDS is used internally in nsec3VerifyNoData.
// See https://datatracker.ietf.org/doc/html/rfc5155#section-8.6
func nsec3ValidateNoDataDS(qname string, nsec3RRSet []dns.RR) (err error) {
	qnameMatchingNSEC3 := nsec3FindMatching(qname, nsec3RRSet)
	if qnameMatchingNSEC3 != nil {
		err = verifyNoDataNsecxTypesDS("NSEC3", qnameMatchingNSEC3.TypeBitMap)
		if err != nil {
			return fmt.Errorf("for qname %s: %w", qname, err)
		}
		return nil
	}

	// No matching NSEC3 found, first verify the closest encloser proof
	// for qname exists.
	closestEncloser, err := nsec3VerifyClosestEncloserProof(qname, nsec3RRSet)
	if err != nil {
		return fmt.Errorf("for qname %s: "+
			"validating closest encloser proof: %w",
			qname, err)
	}
	nextCloser := getNextCloser(qname, closestEncloser)

	// Verify the NSEC3 covering the next closer name has the Opt-Out bit set.
	nextCloserCoveringNSEC3 := nsec3FindCovering(nextCloser, nsec3RRSet)
	if nextCloserCoveringNSEC3 == nil {
		return fmt.Errorf("for qname %s: %w: "+
			"no NSEC3 covers next closer %s",
			qname, errBogus, nextCloser)
	}

	optOutBitSet := nextCloserCoveringNSEC3.Flags == 1
	if !optOutBitSet {
		return fmt.Errorf("for qname %s: %w: "+
			"NSEC3 covering next closer %s Opt-Out bit %d is not set",
			qname, errBogus, nextCloser, nextCloserCoveringNSEC3.Flags)
	}

	return nil
}

// See https://datatracker.ietf.org/doc/html/rfc5155#section-8.7
func nsec3ValidateNoDataWildcard(qname string, qType uint16,
	nsec3RRSet []dns.RR,
) (err error) {
	// Proof qname does not exist with the closest encloser proof
	closestEncloser, err := nsec3VerifyClosestEncloserProof(qname, nsec3RRSet)
	if err != nil {
		return fmt.Errorf("for qname %s: "+
			"validating closest encloser proof: %w",
			qname, err)
	}

	// Proof the wildcard matching qname exists
	wildcardName := "*." + closestEncloser
	err = nsec3RRSetHasMatchingWithoutTypes(nsec3RRSet,
		wildcardName, qType, dns.TypeCNAME)
	if err != nil {
		return fmt.Errorf("for qname %s: %w", qname, err)
	}

	return nil
}

// See https://datatracker.ietf.org/doc/html/rfc5155#section-8.8
func nsec3ValidateWildcard(qname string, nsec3RRSet []dns.RR) (err error) {
	candidateClosestEncloser, err := nsec3VerifyClosestEncloserProof(qname, nsec3RRSet)
	if err != nil {
		return fmt.Errorf("for qname %s: "+
			"validating closest encloser proof: %w",
			qname, err)
	}
	// This closest encloser is the immediate ancestor to the
	// generating wildcard.

	// Validators MUST verify that there is an NSEC3 RR that covers the
	// "next closer" name to QNAME present in the response.  This proves
	// that QNAME itself did not exist and that the correct wildcard was
	// used to generate the response.
	nextCloser := getNextCloser(qname, candidateClosestEncloser)
	nextCloserCoveringNSEC3 := nsec3FindCovering(nextCloser, nsec3RRSet)
	if nextCloserCoveringNSEC3 != nil {
		return nil
	}

	return fmt.Errorf("for qname %s: %w: "+
		"no NSEC3 covers next closer %s",
		qname, errBogus, nextCloser)
}

// The delegationName argument is the owner name of the NS RRSet in the
// authority section of the response.
// See https://datatracker.ietf.org/doc/html/rfc5155#section-8.9
//
//nolint:unused
func nsec3ValidateReferralsToUnsignedSubzones(qname, delegationName string,
	nsec3RRSet []dns.RR,
) (err error) {
	matchingNSEC3 := nsec3FindMatching(qname, nsec3RRSet)
	if matchingNSEC3 != nil {
		var hasNS bool
		for _, nsec3Type := range matchingNSEC3.TypeBitMap {
			switch nsec3Type {
			case dns.TypeNS:
				// This implies the absence of a DNAME type
				hasNS = true
			case dns.TypeDS:
				return fmt.Errorf("for qname %s and delegation name %s: %w: "+
					"NSEC3 matching the delegation name contains DS type",
					qname, delegationName, errBogus)
			case dns.TypeSOA:
				return fmt.Errorf("for qname %s and delegation name %s: %w: "+
					"NSEC3 matching the delegation name contains SOA type",
					qname, delegationName, errBogus)
			}
		}

		if !hasNS {
			return fmt.Errorf("for qname %s and delegation name %s: %w: "+
				"NSEC3 matching the delegation name does not contain NS type",
				qname, delegationName, errBogus)
		}

		return nil
	}

	// No NSEC3 matching the delegation name found
	closestEncloser, err := nsec3VerifyClosestEncloserProof(
		delegationName, nsec3RRSet)
	if err != nil {
		return fmt.Errorf("for qname %s and delegation name %s: "+
			"validating closest encloser proof: %w",
			qname, delegationName, err)
	}

	nextCloser := getNextCloser(delegationName, closestEncloser)
	nextCloserCoveringNSEC3 := nsec3FindCovering(nextCloser, nsec3RRSet)
	if nextCloserCoveringNSEC3 == nil {
		return fmt.Errorf("for qname %s and delegation name %s: %w: "+
			"no NSEC3 covers next closer %s",
			qname, delegationName, errBogus, nextCloser)
	}

	optOutBitSet := nextCloserCoveringNSEC3.Flags == 1
	if optOutBitSet {
		return nil
	}
	return fmt.Errorf("for qname %s and delegation name %s: %w: "+
		"NSEC3 covering next closer %s Opt-Out bit %d is not set",
		qname, delegationName, errBogus, nextCloser, nextCloserCoveringNSEC3.Flags)
}

// nsec3VerifyClosestEncloserProof validates a closest encloser proof,
// and returns the closest encloser name if the proof is valid.
// If the proof is not valid, an error is returned.
// For such proof to be valid, the longest name X must be found such that:
//   - X is an ancestor of qname that is matched by an NSEC3 RR
//   - the name one label longer than X (ancestor of qname or equal to qname)
//     is covered by an NSEC3 RR.
//
// See https://datatracker.ietf.org/doc/html/rfc5155#section-8.3
// The implementation is based on the pseudo code from the RFC.
func nsec3VerifyClosestEncloserProof(qname string, nsec3RRSet []dns.RR) (
	closestEncloser string, err error,
) {
	sname := qname

	for {
		var matchingNSEC3 *dns.NSEC3
		snameCovered := false
		for _, nsec3RR := range nsec3RRSet {
			nsec3 := mustRRToNSEC3(nsec3RR)

			if nsec3.Cover(sname) {
				snameCovered = true
			}

			if nsec3.Match(sname) {
				matchingNSEC3 = nsec3
			}
		}

		if matchingNSEC3 != nil {
			if !snameCovered {
				return "", fmt.Errorf("%w: sname %s matched but not covered",
					errBogus, sname)
			}
			closestEncloser = sname

			// The DNAME type bit must not be set and the NS type bit may
			// only be set if the SOA type bit is set.
			// If this is not the case, it would be an indication that an attacker
			// is using them to falsely deny the existence of RRs for which the
			// server is not authoritative.
			var hasNS, hasSOA bool
			for _, nsec3Type := range matchingNSEC3.TypeBitMap {
				switch nsec3Type {
				case dns.TypeDNAME:
					return "", fmt.Errorf("%w: NSEC3 of closest encloser %s "+
						"contains the DNAME type", errBogus, sname)
				case dns.TypeNS:
					hasNS = true
				case dns.TypeSOA:
					hasSOA = true
				}
			}
			if hasNS && !hasSOA {
				return "", fmt.Errorf("%w: NSEC3 of closest encloser %s "+
					"contains the NS type but not the SOA type", errBogus, sname)
			}

			return closestEncloser, nil
		}

		const offset = 0
		i, end := dns.NextLabel(sname, offset)
		if end {
			return "", fmt.Errorf("%w: sname reached the last label already", errBogus)
		}
		sname = sname[i:]
	}
}

// getNextCloser returns the "next closer" name of qname given a closest
// encloser name.
// For example with qname="a.b.example.com." and closestEncloser=".com.",
// then nextCloser="example.com.".
func getNextCloser(qname, closestEncloser string) (nextCloser string) {
	closestEncloserLabelsCount := dns.CountLabel(closestEncloser)

	nextCloserStartIndex, startOvershoot := dns.PrevLabel(qname, closestEncloserLabelsCount+1)
	if startOvershoot {
		panic("start overshoot should not happen")
	}
	nextCloser = qname[nextCloserStartIndex:]

	return nextCloser
}

// nsec3RRSetHasMatchingWithoutTypes returns an error if:
// - there is no NSEC3 matching matchName
// - the NSEC3 matching matchName contains one of the notTypes.
func nsec3RRSetHasMatchingWithoutTypes(nsec3RRSet []dns.RR,
	matchName string, notTypes ...uint16,
) (err error) {
	matchingNSEC3 := nsec3FindMatching(matchName, nsec3RRSet)
	if matchingNSEC3 == nil {
		return fmt.Errorf("%w: no NSEC3 matching %s",
			errBogus, matchName)
	}

	for _, nsec3Type := range matchingNSEC3.TypeBitMap {
		for _, notType := range notTypes {
			if nsec3Type != notType {
				continue
			}
			return fmt.Errorf("%w: NSEC3 matching %s contains type %s",
				errBogus, matchName, dns.TypeToString[notType])
		}
	}

	return nil
}

func nsec3FindMatching(qname string, nsec3RRSet []dns.RR) (
	matchingNSEC3 *dns.NSEC3,
) {
	for _, nsec3RR := range nsec3RRSet {
		nsec3 := mustRRToNSEC3(nsec3RR)
		if nsec3.Match(qname) {
			return nsec3
		}
	}
	return nil
}

func nsec3FindCovering(qname string, nsec3RRSet []dns.RR) (
	coveringNSEC3 *dns.NSEC3,
) {
	for _, nsec3RR := range nsec3RRSet {
		nsec3 := mustRRToNSEC3(nsec3RR)
		if nsec3.Cover(qname) {
			return nsec3
		}
	}
	return nil
}
