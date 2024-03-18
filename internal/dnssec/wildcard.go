package dnssec

import (
	"errors"
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

// extractWildcardExpansion returns an empty string if no wildcard expansion
// is found, otherwise it returns the wildcard name in the format "*.domain.tld.".
func extractWildcardExpansion(signedRRSets []dnssecRRSet) (wildcardName string) {
	// TODO simplify this once tested live
	var expandedQtype uint16 // TODO remove safety check
	for _, signedRRSet := range signedRRSets {
		for _, rrSig := range signedRRSet.rrSigs {
			if !isRRSigForWildcard(rrSig) {
				continue
			}

			labels := dns.SplitDomainName(rrSig.Hdr.Name)
			newWildcardName := dns.Fqdn("*." + strings.Join(labels[len(labels)-int(rrSig.Labels):], "."))
			if wildcardName != "" && wildcardName != newWildcardName {
				globalDebugLogger.Errorf("wildcard expanded multiple names: %s and %s",
					wildcardName, newWildcardName)
			}
			wildcardName = newWildcardName

			if expandedQtype != dns.TypeNone && expandedQtype != rrSig.TypeCovered {
				globalDebugLogger.Errorf("wildcard expanded multiple types: %s and %s",
					dns.TypeToString[expandedQtype], dns.TypeToString[rrSig.TypeCovered])
			}
			expandedQtype = rrSig.TypeCovered
		}
	}

	return wildcardName
}

var (
	ErrNSECxMissing = errors.New("NSEC or NSEC3 record missing")
)

// For wildcard considerations in positive responses, see:
// - https://datatracker.ietf.org/doc/html/rfc2535#section-5.3
// - https://datatracker.ietf.org/doc/html/rfc4035#section-5.3.4
// - https://datatracker.ietf.org/doc/html/rfc4035#section-3.1.3.3
func validateWildcardExpansion(expandedQname string,
	authoritySection []dnssecRRSet) (err error) {
	nsec3RRs, wildcard := extractNSEC3s(authoritySection)
	if len(nsec3RRs) > 0 {
		nsec3RRs, err = nsec3InitialChecks(nsec3RRs)
		if err != nil {
			return fmt.Errorf("initial NSEC3 checks: %w", err)
		}
		globalDebugLogger.Infof("validating wildcard expansion for %s: "+
			"NSEC3 wildcard: %t", expandedQname, wildcard)
		return nsec3ValidateWildcard(expandedQname, nsec3RRs)
	}

	nsecRRs := extractNSECs(authoritySection)
	if len(nsecRRs) > 0 {
		return nsecValidateNxDomain(expandedQname, nsecRRs)
	}

	return fmt.Errorf("%w", ErrNSECxMissing)
}
