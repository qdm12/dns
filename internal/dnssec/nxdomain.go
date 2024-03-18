package dnssec

import (
	"errors"
	"fmt"

	"github.com/miekg/dns"
)

var (
	ErrRRSigWildcardUnexpected = errors.New("RRSIG for a wildcard is unexpected")
)

func validateNxDomain(qname string, authoritySection []dnssecRRSet,
	keyTagToDNSKey map[uint16]*dns.DNSKEY) (err error) {
	err = verifyRRSetsRRSig(authoritySection, keyTagToDNSKey)
	if err != nil {
		return fmt.Errorf("verifying RRSIGs: %w", err)
	}

	nsec3RRs, wildcard := extractNSEC3s(authoritySection)
	if wildcard {
		return fmt.Errorf("for NXDOMAIN response for %s: NSEC3: %w",
			qname, ErrRRSigWildcardUnexpected)
	} else if len(nsec3RRs) > 0 {
		nsec3RRs, err = nsec3InitialChecks(nsec3RRs)
		if err != nil {
			return fmt.Errorf("initial NSEC3 checks: %w", err)
		}
		return nsec3ValidateNxDomain(qname, nsec3RRs)
	}

	nsecRRs := extractNSECs(authoritySection)
	if len(nsecRRs) > 0 {
		return nsecValidateNxDomain(qname, nsecRRs)
	}

	return fmt.Errorf("for %s: %w: no NSEC or NSEC3 record found",
		qname, ErrBogus)
}
