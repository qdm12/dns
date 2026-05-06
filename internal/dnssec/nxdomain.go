package dnssec

import (
	"errors"
	"fmt"
)

var errRRSigWildcardUnexpected = errors.New("RRSIG for a wildcard is unexpected")

func validateNxDomain(qname string, authoritySection []dnssecRRSet,
	keyTagToDNSKeys dnsKeysByTag,
) (err error) {
	err = verifyRRSetsRRSig(authoritySection, keyTagToDNSKeys)
	if err != nil {
		return fmt.Errorf("verifying RRSIGs: %w", err)
	}

	nsec3RRs, wildcard := extractNSEC3s(authoritySection)
	if wildcard {
		return fmt.Errorf("for NXDOMAIN response for %s: NSEC3: %w",
			qname, errRRSigWildcardUnexpected)
	} else if len(nsec3RRs) > 0 {
		nsec3RRs, err = nsec3InitialChecks(nsec3RRs, keyTagToDNSKeys)
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
		qname, errBogus)
}
