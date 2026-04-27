package dnssec

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

func mustRRToNSEC(rr dns.RR) (nsec *dns.NSEC) {
	nsec, ok := rr.(*dns.NSEC)
	if !ok {
		panic(fmt.Sprintf("RR is of type %T and not of type *dns.NSEC", rr))
	}
	return nsec
}

// extractNSECs returns the NSEC RRs found in the NSEC
// signed RRSet from the slice of signed RRSets.
func extractNSECs(rrSets []dnssecRRSet) (nsecs []dns.RR) {
	for _, rrSet := range rrSets {
		if rrSet.qtype() == dns.TypeNSEC {
			return rrSet.rrSet
		}
	}
	return nil
}

func nsecValidateNxDomain(qname string, nsecRRSet []dns.RR) (err error) {
	for _, nsecRR := range nsecRRSet {
		nsec := mustRRToNSEC(nsecRR)
		if nsecCoversZone(qname, nsec.Hdr.Name, nsec.NextDomain) {
			return nil
		}
	}

	return fmt.Errorf("for qname %s: %w: "+
		"no NSEC covering qname found",
		qname, errBogus)
}

func nsecValidateNoData(qname string, qType uint16,
	nsecRRSet []dns.RR,
) (err error) {
	if qType == dns.TypeDS {
		return nsecValidateNoDataDS(qname, nsecRRSet)
	}

	var qnameMatchingNSEC *dns.NSEC
	for _, nsecRR := range nsecRRSet {
		nsec := mustRRToNSEC(nsecRR)
		if nsecMatchesQname(nsec, qname) {
			qnameMatchingNSEC = nsec
			break
		}
	}

	if qnameMatchingNSEC == nil {
		return fmt.Errorf("for zone %s and type %s: %w: "+
			"no NSEC matching qname found",
			qname, dns.TypeToString[qType], errBogus)
	}

	for _, nsecType := range qnameMatchingNSEC.TypeBitMap {
		switch nsecType {
		case qType:
			return fmt.Errorf("for qname %s and type %s: %w: "+
				"qtype contained in NSEC",
				qname, dns.TypeToString[qType], errBogus)
		case dns.TypeCNAME:
			// Per RFC 4034 §4.4: A CNAME RR and other RRs cannot coexist at the same owner name.
			// Therefore, if NSEC type bitmap contains CNAME for a qtype query that is not CNAME,
			// the NSEC is invalid (it claims both CNAME and qtype exist at the same owner).
			return fmt.Errorf("for qname %s and type %s: %w: "+
				"NSEC contains both CNAME and qtype (RFC 4034 violation)",
				qname, dns.TypeToString[qType], errBogus)
		}
	}

	return nil
}

func nsecValidateNoDataDS(qname string, nsecRRSet []dns.RR) (err error) {
	var qnameMatchingNSEC *dns.NSEC
	for _, nsecRR := range nsecRRSet {
		nsec := mustRRToNSEC(nsecRR)
		if nsecMatchesQname(nsec, qname) {
			qnameMatchingNSEC = nsec
			break
		}
	}

	if qnameMatchingNSEC == nil {
		return fmt.Errorf("for qname %s: %w: "+
			"no NSEC matching qname found",
			qname, errBogus)
	}

	err = verifyNoDataNsecxTypesDS("NSEC", qnameMatchingNSEC.TypeBitMap)
	if err != nil {
		return fmt.Errorf("for qname %s: %w",
			qname, err)
	}
	return nil
}

// nsecMatchesQname returns true if the NSEC owner name is equal
// to the qname or if the NSEC owner name is a wildcard name parent
// of qname.
func nsecMatchesQname(nsec *dns.NSEC, qname string) bool {
	return nsec.Hdr.Name == qname || (strings.HasPrefix(nsec.Hdr.Name, "*.") &&
		dns.IsSubDomain(nsec.Hdr.Name[2:], qname))
}

// nsecCoversZone returns true if the zone is within the OPEN interval
// delimited by the nsecOwner and the nsecNext FQDNs given.
// The comparison follows canonical DNS name ordering from RFC 4034 section 6.1,
// matching the approach taken by Unbound's dname_canonical_compare.
func nsecCoversZone(zone, nsecOwner, nsecNext string) (ok bool) {
	if zone == nsecOwner || zone == nsecNext {
		return false
	}

	ownerToNext := canonicalNameCompare(nsecOwner, nsecNext)
	ownerToZone := canonicalNameCompare(nsecOwner, zone)
	zoneToNext := canonicalNameCompare(zone, nsecNext)

	switch {
	case ownerToNext < 0:
		// Regular interval: (owner, next)
		return ownerToZone < 0 && zoneToNext < 0
	case ownerToNext > 0:
		// Wrapped interval across the canonical ordering end:
		// zone in (owner, +inf) U (-inf, next)
		return ownerToZone < 0 || zoneToNext < 0
	default:
		// owner == next means the NSEC covers the full canonical range,
		// except owner/next itself which is filtered above.
		return true
	}
}

// canonicalNameCompare compares two domain names using RFC 4034 canonical
// ordering, label-by-label from the zone apex towards the left-most label.
// Returns -1 if a < b, 0 if a == b, +1 if a > b.
func canonicalNameCompare(a, b string) int {
	aLabels := dns.SplitDomainName(strings.ToLower(a))
	bLabels := dns.SplitDomainName(strings.ToLower(b))

	minLabels := min(len(aLabels), len(bLabels))
	for i := 1; i <= minLabels; i++ {
		al := aLabels[len(aLabels)-i]
		bl := bLabels[len(bLabels)-i]

		if al == bl {
			continue
		}

		// RFC 4034 canonical label ordering compares by bytes and if equal
		// up to min length, the shorter label sorts first.
		minLen := min(len(al), len(bl))
		for j := range minLen {
			if al[j] < bl[j] {
				return -1
			}
			if al[j] > bl[j] {
				return 1
			}
		}

		if len(al) < len(bl) {
			return -1
		}
		return 1
	}

	if len(aLabels) < len(bLabels) {
		return -1
	}
	if len(aLabels) > len(bLabels) {
		return 1
	}
	return 0
}
