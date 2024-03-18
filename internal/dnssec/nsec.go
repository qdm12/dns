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
		qname, ErrBogus)
}

func nsecValidateNoData(qname string, qType uint16,
	nsecRRSet []dns.RR) (err error) {
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
			qname, dns.TypeToString[qType], ErrBogus)
	}

	for _, nsecType := range qnameMatchingNSEC.TypeBitMap {
		switch nsecType {
		case qType:
			return fmt.Errorf("for qname %s and type %s: %w: "+
				"qtype contained in NSEC",
				qname, dns.TypeToString[qType], ErrBogus)
		case dns.TypeCNAME: // TODO check this is invalid
			return fmt.Errorf("for qname %s and type %s: %w: "+
				"CNAME contained in NSEC",
				qname, dns.TypeToString[qType], ErrBogus)
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
			qname, ErrBogus)
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
// TODO improve inspiring from
// https://github.com/NLnetLabs/unbound/blob/master/util/data/dname.c#L802
func nsecCoversZone(zone, nsecOwner, nsecNext string) (ok bool) {
	if zone == nsecOwner || zone == nsecNext {
		return false
	}

	zoneLabels := dns.SplitDomainName(zone)
	nsecOwnerLabels := dns.SplitDomainName(nsecOwner)

	if len(zoneLabels) < len(nsecOwnerLabels) {
		// zone is shorter than NSEC owner, so it cannot be covered
		return false
	}

	for i := range nsecOwnerLabels {
		zoneLabel := zoneLabels[len(zoneLabels)-1-i]
		nsecOwnerLabel := nsecOwnerLabels[len(nsecOwnerLabels)-1-i]
		if nsecOwnerLabel == "*" {
			// wildcard NSEC owner containing zone
			return true
		} else if zoneLabel < nsecOwnerLabel {
			return false
		}
	}

	nsecNextLabels := dns.SplitDomainName(nsecNext)
	if len(zoneLabels) < len(nsecNextLabels) {
		// zone is shorter than NSEC next, so it cannot be covered
		return false
	}

	minLabelsCount := min(len(zoneLabels), len(nsecNextLabels))
	for i := 0; i < minLabelsCount; i++ {
		zoneLabel := zoneLabels[len(zoneLabels)-1-i]
		nsecNextLabel := nsecNextLabels[len(nsecNextLabels)-1-i]
		if zoneLabel > nsecNextLabel {
			return false
		}
	}

	// Zone and next domain have the same labels for the first
	// minLabelsCount labels, and zone != next, so zone is within
	// the interval delimited by owner and next.
	return true
}
