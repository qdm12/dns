package dnssec

import (
	"errors"
	"fmt"

	"github.com/miekg/dns"
)

// dnssecRRSet is a possibly signed RRSet for a certain
// owner, type and class, containing at least one or more
// RRs and zero or more RRSigs.
// If the RRSet is unsigned, the rrSigs field is a slice
// of length 0.
type dnssecRRSet struct {
	// rrSigs is the slice of RRSIGs for the RRSet.
	// There can be more than one RRSIG, for example:
	// dig +dnssec -t A xyzzy14.sdsmt.edu. @1.1.1.1
	// returns 2 RRSIGs for the SOA authority section RRSet.
	rrSigs []*dns.RRSIG
	// rrSet cannot be empty.
	rrSet []dns.RR
}

func (d dnssecRRSet) qtype() uint16 {
	return d.rrSet[0].Header().Rrtype
}

func (d dnssecRRSet) ownerAndType() string {
	return d.rrSet[0].Header().Name + " " +
		dns.TypeToString[d.rrSet[0].Header().Rrtype]
}

func dnssecRRSetsToRRs(rrSets []dnssecRRSet, ignoreTypes ...uint16) (rrs []dns.RR) {
	if len(rrSets) == 0 {
		return nil
	}

	ignoreTypesMap := make(map[uint16]struct{}, len(ignoreTypes))
	for _, ignoreType := range ignoreTypes {
		ignoreTypesMap[ignoreType] = struct{}{}
	}

	minRRSetSize := len(rrSets) // 1 RR per owner, type and class
	rrs = make([]dns.RR, 0, minRRSetSize)
	for _, rrSet := range rrSets {
		for _, rr := range rrSet.rrSet {
			rrType := rr.Header().Rrtype
			_, ignore := ignoreTypesMap[rrType]
			if ignore {
				continue
			}
			rrs = append(rrs, rr)
		}

		_, rrSigIgnored := ignoreTypesMap[dns.TypeRRSIG]
		if rrSigIgnored {
			continue
		}

		for _, rrSig := range rrSet.rrSigs {
			_, ignored := ignoreTypesMap[rrSig.TypeCovered]
			if ignored {
				continue
			}
			rrs = append(rrs, rrSig)
		}
	}

	return rrs
}

var (
	ErrRRSetsMissing       = errors.New("no RRSet")
	ErrRRSetsMultiple      = errors.New("multiple RRSets")
	ErrRRSetTypeUnexpected = errors.New("RRSet type unexpected")
)

func dnssecRRSetsIsSingleOfType(rrSets []dnssecRRSet, qType uint16) (err error) {
	switch {
	case len(rrSets) == 0:
		return fmt.Errorf("%w", ErrRRSetsMissing)
	case len(rrSets) == 1:
	default:
		return fmt.Errorf("%w: received %d RRSets instead of 1",
			ErrRRSetsMultiple, len(rrSets))
	}

	rrSetType := rrSets[0].qtype()
	if rrSetType != qType {
		return fmt.Errorf("%w: received %s RRSet instead of %s",
			ErrRRSetTypeUnexpected, dns.TypeToString[rrSetType],
			dns.TypeToString[qType])
	}

	return nil
}

func removeFromRRSet(rrSet []dns.RR, typesToRemove ...uint16) (filtered []dns.RR) {
	if len(rrSet) == 0 {
		return nil
	}

	filtered = make([]dns.RR, 0, len(rrSet))
	for _, rr := range rrSet {
		rrType := rr.Header().Rrtype
		for _, rrTypeToRemove := range typesToRemove {
			if rrType == rrTypeToRemove {
				continue
			}
		}
		filtered = append(filtered, rr)
	}
	return filtered
}
