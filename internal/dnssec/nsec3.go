package dnssec

import (
	"errors"
	"fmt"
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
// wildcard as true if the NSEC3 RRSet RRSig is for a wildcard.
func extractNSEC3s(rrSets []dnssecRRSet) (
	rrs []dns.RR, wildcard bool) {
	rrs = make([]dns.RR, 0, len(rrSets))
	for _, rrSet := range rrSets {
		if rrSet.qtype() == dns.TypeNSEC3 {
			if !wildcard {
				for _, rrSig := range rrSet.rrSigs {
					if isRRSigForWildcard(rrSig) {
						wildcard = true
						break
					}
				}
			}
			rrs = append(rrs, rrSet.rrSet...)
		}
	}
	return rrs, wildcard
}

var (
	ErrNSEC3RRSetDifferentHashTypes  = errors.New("NSEC3 RRSet contains different hash types")
	ErrNSEC3RRSetDifferentIterations = errors.New("NSEC3 RRSet contains different iterations")
	ErrNSEC3RRSetDifferentSalts      = errors.New("NSEC3 RRSet contains different salts")
)

func nsec3InitialChecks(nsec3RRSet []dns.RR) (sanitizedNSEC3RRSet []dns.RR, err error) {
	sanitizedNSEC3RRSet = make([]dns.RR, 0, len(nsec3RRSet))

	const usualCapacity = 1
	hashTypes := make(map[uint8]struct{}, usualCapacity)
	iterations := make(map[uint16]struct{}, usualCapacity)
	salts := make(map[string]struct{}, usualCapacity)

	for _, nsec3RR := range nsec3RRSet {
		nsec3 := mustRRToNSEC3(nsec3RR)

		// Only accept supported hash type
		// https://datatracker.ietf.org/doc/html/rfc5155#section-8.1
		if !isOneOf(nsec3.Hash, dns.SHA1) {
			continue
		}

		// Flag field must be zero or one (opt-out).
		// https://datatracker.ietf.org/doc/html/rfc5155#section-8.2
		if !isOneOf(nsec3.Flags, 0, 1) {
			continue
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
		return nil, fmt.Errorf("%w: %s", ErrNSEC3RRSetDifferentHashTypes,
			hashesToString(maps.Keys(hashTypes)))
	case len(iterations) > 1:
		return nil, fmt.Errorf("%w: %s", ErrNSEC3RRSetDifferentIterations,
			integersToString(maps.Keys(iterations)))
	case len(salts) > 1:
		return nil, fmt.Errorf("%w: %s", ErrNSEC3RRSetDifferentSalts,
			strings.Join(maps.Keys(salts), ", "))
	}

	return sanitizedNSEC3RRSet, nil
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
			qname, ErrBogus, wildcardName)
	}

	return nil
}

// nsec3ValidateNoData validates a no data response for a given QTYPE.
// See https://datatracker.ietf.org/doc/html/rfc5155#section-8.5
// and https://datatracker.ietf.org/doc/html/rfc5155#section-8.6
func nsec3ValidateNoData(qname string, qType uint16,
	nsec3RRSet []dns.RR) (err error) {
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
			qname, ErrBogus, nextCloser)
	}

	optOutBitSet := nextCloserCoveringNSEC3.Flags == 1
	if !optOutBitSet {
		return fmt.Errorf("for qname %s: %w: "+
			"NSEC3 covering next closer %s Opt-Out bit %d is not set",
			qname, ErrBogus, nextCloser, nextCloserCoveringNSEC3.Flags)
	}

	return nil
}

// See https://datatracker.ietf.org/doc/html/rfc5155#section-8.7
func nsec3ValidateNoDataWildcard(qname string, qType uint16,
	nsec3RRSet []dns.RR) (err error) {
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
		qname, ErrBogus, nextCloser)
}

// The delegationName argument is the owner name of the NS RRSet in the
// authority section of the response.
// See https://datatracker.ietf.org/doc/html/rfc5155#section-8.9
func nsec3ValidateReferralsToUnsignedSubzones(qname, delegationName string,
	nsec3RRSet []dns.RR) (err error) {
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
					qname, delegationName, ErrBogus)
			case dns.TypeSOA:
				return fmt.Errorf("for qname %s and delegation name %s: %w: "+
					"NSEC3 matching the delegation name contains SOA type",
					qname, delegationName, ErrBogus)
			}
		}

		if !hasNS {
			return fmt.Errorf("for qname %s and delegation name %s: %w: "+
				"NSEC3 matching the delegation name does not contain NS type",
				qname, delegationName, ErrBogus)
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
			qname, delegationName, ErrBogus, nextCloser)
	}

	optOutBitSet := nextCloserCoveringNSEC3.Flags == 1
	if optOutBitSet {
		return nil
	}
	return fmt.Errorf("for qname %s and delegation name %s: %w: "+
		"NSEC3 covering next closer %s Opt-Out bit %d is not set",
		qname, delegationName, ErrBogus, nextCloser, nextCloserCoveringNSEC3.Flags)
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
	closestEncloser string, err error) {
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
					ErrBogus, sname)
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
						"contains the DNAME type", ErrBogus, sname)
				case dns.TypeNS:
					hasNS = true
				case dns.TypeSOA:
					hasSOA = true
				}
			}
			if hasNS && !hasSOA {
				return "", fmt.Errorf("%w: NSEC3 of closest encloser %s "+
					"contains the NS type but not the SOA type", ErrBogus, sname)
			}

			return closestEncloser, nil
		}

		const offset = 0
		i, end := dns.NextLabel(sname, offset)
		if end {
			return "", fmt.Errorf("%w: sname reached the last label already", ErrBogus)
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
	qnameLabelsCount := dns.CountLabel(qname)

	// Double check the qname is two labels longer than the closest encloser.
	// TODO eventual remove check
	if qnameLabelsCount < closestEncloserLabelsCount+1 {
		panic(fmt.Sprintf("qname %s is not at least one label longer than closest encloser %s",
			qname, closestEncloser))
	}

	nextCloserStartIndex, startOvershoot := dns.PrevLabel(qname, closestEncloserLabelsCount+1)
	if startOvershoot {
		panic("start overshoot should not happen")
	}
	nextCloser = qname[nextCloserStartIndex:]

	return nextCloser
}

// nsec3RRSetHasMatchingWithoutTypes returns an error if:
// - there is no NSEC3 matching matchName
// - the NSEC3 matching matchName contains one of the notTypes
func nsec3RRSetHasMatchingWithoutTypes(nsec3RRSet []dns.RR,
	matchName string, notTypes ...uint16) (err error) {
	matchingNSEC3 := nsec3FindMatching(matchName, nsec3RRSet)
	if matchingNSEC3 == nil {
		return fmt.Errorf("%w: no NSEC3 matching %s",
			ErrBogus, matchName)
	}

	for _, nsec3Type := range matchingNSEC3.TypeBitMap {
		for _, notType := range notTypes {
			if nsec3Type != notType {
				continue
			}
			return fmt.Errorf("%w: NSEC3 matching %s contains type %s",
				ErrBogus, matchName, dns.TypeToString[notType])
		}
	}

	return nil
}

func nsec3FindMatching(qname string, nsec3RRSet []dns.RR) (
	matchingNSEC3 *dns.NSEC3) {
	for _, nsec3RR := range nsec3RRSet {
		nsec3 := mustRRToNSEC3(nsec3RR)
		if nsec3.Match(qname) {
			return nsec3
		}
	}
	return nil
}

func nsec3FindCovering(qname string, nsec3RRSet []dns.RR) (
	coveringNSEC3 *dns.NSEC3) {
	for _, nsec3RR := range nsec3RRSet {
		nsec3 := mustRRToNSEC3(nsec3RR)
		if nsec3.Cover(qname) {
			return nsec3
		}
	}
	return nil
}
