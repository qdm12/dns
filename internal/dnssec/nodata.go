package dnssec

import (
	"fmt"

	"github.com/miekg/dns"
)

// Note: validateNoData works also for the qtype DS since
// the implementations of nsec3ValidateNoData and
// nsecValidateNoData take care of redirecting to the
// DS specific validation functions, but preferably use
// validateNoDataDS for the qtype DS.
func validateNoData(qname string, qtype uint16,
	authoritySection []dnssecRRSet,
	keyTagToDNSKey map[uint16]*dns.DNSKEY) (err error) {
	err = verifyRRSetsRRSig(nil, authoritySection, keyTagToDNSKey)
	if err != nil {
		return fmt.Errorf("verifying RRSIGs: %w", err)
	}

	nsec3RRs, wildcard := extractNSEC3s(authoritySection)
	if len(nsec3RRs) > 0 {
		nsec3RRs, err = nsec3InitialChecks(nsec3RRs)
		if err != nil {
			return fmt.Errorf("initial NSEC3 checks: %w", err)
		} else if wildcard {
			return nsec3ValidateNoDataWildcard(qname, qtype, nsec3RRs)
		}
		return nsec3ValidateNoData(qname, qtype, nsec3RRs)
	}

	nsecRRs := extractNSECs(authoritySection)
	if len(nsecRRs) > 0 {
		return nsecValidateNoData(qname, qtype, nsecRRs)
	}

	return fmt.Errorf("verifying no data for %s: %w: "+
		"no NSEC or NSEC3 record found",
		nameTypeToString(qname, qtype), ErrBogus)
}

func validateNoDataDS(qname string,
	authoritySection []dnssecRRSet,
	keyTagToDNSKey map[uint16]*dns.DNSKEY) (err error) {
	err = verifyRRSetsRRSig(nil, authoritySection, keyTagToDNSKey)
	if err != nil {
		return fmt.Errorf("verifying RRSIGs: %w", err)
	}

	nsec3RRs, wildcard := extractNSEC3s(authoritySection)
	if len(nsec3RRs) > 0 {
		nsec3RRs, err = nsec3InitialChecks(nsec3RRs)
		if err != nil {
			return fmt.Errorf("initial NSEC3 checks: %w", err)
		} else if wildcard {
			return nsec3ValidateNoDataWildcard(qname, dns.TypeDS, nsec3RRs)
		}
		return nsec3ValidateNoDataDS(qname, nsec3RRs)
	}

	nsecRRs := extractNSECs(authoritySection)
	if len(nsecRRs) > 0 {
		return nsecValidateNoDataDS(qname, nsecRRs)
	}

	return fmt.Errorf("verifying no DS data for %s: %w: "+
		"no NSEC or NSEC3 record found",
		qname, ErrBogus)
}

// See https://datatracker.ietf.org/doc/html/rfc5155#section-8.6
func verifyNoDataNsecxTypesDS(nsecVariant string,
	nsecTypes []uint16) (err error) {
	for _, nsecType := range nsecTypes {
		switch nsecType {
		case dns.TypeSOA:
			return fmt.Errorf("%w: %s contains SOA type"+
				" so is from the child zone and not the parent zone",
				ErrBogus, nsecVariant)
		case dns.TypeDS:
			return fmt.Errorf("%w: %s contains DS type", ErrBogus, nsecVariant)
		case dns.TypeCNAME:
			return fmt.Errorf("%w: %s contains CNAME type",
				ErrBogus, nsecVariant)
		}
	}

	return nil
}
