package dnssec

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/miekg/dns"
)

func mustRRToRRSig(rr dns.RR) (rrSig *dns.RRSIG) {
	rrSig, ok := rr.(*dns.RRSIG)
	if !ok {
		panic(fmt.Sprintf("RR is of type %T and not of type *dns.RRSIG", rr))
	}
	return rrSig
}

func rrSigToOwnerTypeCovered(rrSig *dns.RRSIG) (ownerTypeCovered string) {
	return fmt.Sprintf("RRSIG for owner %s and type %s",
		rrSig.Header().Name, dns.TypeToString[rrSig.TypeCovered])
}

// isRRSigForWildcard returns true if the RRSIG is for a wildcard.
// This is detected by checking if the number of labels in the RRSIG
// owner name is less than the number of labels in the RRSig owner name.
// See https://datatracker.ietf.org/doc/html/rfc7129#section-5.3
func isRRSigForWildcard(rrSig *dns.RRSIG) bool {
	if rrSig == nil {
		return false
	}
	ownerLabelsCount := uint8(dns.CountLabel(rrSig.Hdr.Name)) //nolint:gosec
	return rrSig.Labels < ownerLabelsCount
}

var errRRSigLabels = errors.New("RRSIG labels greater than owner labels")

// See https://datatracker.ietf.org/doc/html/rfc4035#section-5.3.1
func rrsigInitialChecks(rrsig *dns.RRSIG) (err error) {
	rrSetOwner := rrsig.Hdr.Name

	if int(rrsig.Labels) > dns.CountLabel(rrSetOwner) {
		// The number of labels in the RRset owner name MUST be greater than
		// or equal to the value in the RRSIG RR's Labels field.
		return fmt.Errorf("for %s: %w: RRSig labels field is %d and owner is %d labels",
			rrSigToOwnerTypeCovered(rrsig), errRRSigLabels,
			rrsig.Labels, dns.CountLabel(rrSetOwner))
	}

	return nil
}

func verifyRRSetsRRSig(answerRRSets []dnssecRRSet, keyTagToDNSKeys dnsKeysByTag) (err error) {
	budget := newRRSIGValidationBudget()
	for _, signedRRSet := range answerRRSets {
		err = verifyRRSetRRSigs(signedRRSet.rrSet,
			signedRRSet.rrSigs, keyTagToDNSKeys, budget)
		if err != nil {
			return err
		}
	}

	return nil
}

func verifyRRSetRRSigs(rrSet []dns.RR, rrSigs []*dns.RRSIG,
	keyTagToDNSKeys dnsKeysByTag, budget *rrsigValidationBudget,
) (
	err error,
) {
	switch {
	case len(rrSet) == 0 || len(rrSigs) == 0:
		panic("no rrs or rrsigs")
	case len(rrSigs) > maxRRSIGValidationsPerRRSet:
		return fmt.Errorf("%w: got %d rrsigs above the limit of %d",
			errRRSIGValidationRRSetBudgetExceeded,
			len(rrSigs), maxRRSIGValidationsPerRRSet)
	case len(rrSigs) == 1:
		return verifyRRSetRRSig(rrSet, rrSigs[0], keyTagToDNSKeys, budget)
	}

	// Multiple RRSIGs for the same RRSet, sort them by algorithm preference
	// and try each one until one succeeds. This is rather undocumented,
	// but one signature verified should be enough to validate the RRSet,
	// even if other signatures fail to verify successfully.
	sortRRSIGsByAlgo(rrSigs)

	errs := new(joinedErrors)
	for _, rrSig := range rrSigs {
		if err = checkRRSigAlgorithm(rrSig); err != nil {
			errs.add(err)
			continue
		}

		err = budget.consume()
		if err != nil {
			return err
		}

		if !rrSig.ValidityPeriod(time.Now()) {
			errs.add(fmt.Errorf("%w", errRRSigExpired))
			continue
		}

		err = checkRRSigSignerName(rrSig, keyTagToDNSKeys)
		if err != nil {
			errs.add(err)
			continue
		}

		matchingDNSKeys := matchingDNSKeysForRRSIG(rrSig, keyTagToDNSKeys)
		if len(matchingDNSKeys) == 0 {
			errs.add(fmt.Errorf("%w: for signer %s, algorithm %d and key tag %d",
				errRRSigDNSKey, rrSig.SignerName, rrSig.Algorithm, rrSig.KeyTag))
			continue
		}

		var verified bool
		for _, dnsKey := range matchingDNSKeys {
			err = rrSig.Verify(dnsKey, rrSet)
			if err == nil {
				verified = true
				break
			}
			errs.add(err)
		}
		if verified {
			return nil
		}
	}

	return fmt.Errorf("%d RRSIGs failed to validate the RRSet: %w",
		len(rrSigs), errs)
}

var (
	errRRSigDNSKey               = errors.New("DNSKEY not found")
	errRRSigSignerName           = errors.New("RRSIG signer name is not zone apex")
	errRRSigExpired              = errors.New("RRSIG has expired")
	errRRSigForbiddenAlgorithm   = errors.New("RRSIG algorithm is forbidden by RFC 8624")
	errRRSigUnsupportedAlgorithm = errors.New("RRSIG algorithm is not supported")
)

func matchingDNSKeysForRRSIG(rrSig *dns.RRSIG, keyTagToDNSKeys dnsKeysByTag) []*dns.DNSKEY {
	dnsKeys := keyTagToDNSKeys[rrSig.KeyTag]
	if len(dnsKeys) == 0 {
		return nil
	}

	matches := make([]*dns.DNSKEY, 0, len(dnsKeys))
	for _, dnsKey := range dnsKeys {
		if !strings.EqualFold(dnsKey.Header().Name, rrSig.SignerName) {
			continue
		}
		if dnsKey.Algorithm != rrSig.Algorithm {
			continue
		}
		matches = append(matches, dnsKey)
	}

	return matches
}

// checkRRSigAlgorithm returns an error if the RRSIG's algorithm must not
// or cannot be used for validation per RFC 8624 section 3.1.
// This enforces the algorithm policy rather than merely preferring stronger
// algorithms: MustNot algorithms are hard-rejected and Unknown algorithms
// are skipped (RFC 4035 section 5.3.1: treat as if not present).
func checkRRSigAlgorithm(rrSig *dns.RRSIG) error {
	switch algoIDToPreference(rrSig.Algorithm) {
	case algoPreferenceMustNot:
		return fmt.Errorf("%w: %s",
			errRRSigForbiddenAlgorithm, dns.AlgorithmToString[rrSig.Algorithm])
	case algoPreferenceUnknown:
		return fmt.Errorf("%w: algorithm %d",
			errRRSigUnsupportedAlgorithm, rrSig.Algorithm)
	}
	return nil
}

func verifyRRSetRRSig(rrSet []dns.RR, rrSig *dns.RRSIG,
	keyTagToDNSKeys dnsKeysByTag, budget *rrsigValidationBudget,
) (err error) {
	if err = checkRRSigAlgorithm(rrSig); err != nil {
		return err
	}

	err = budget.consume()
	if err != nil {
		return err
	}

	if !rrSig.ValidityPeriod(time.Now()) {
		return fmt.Errorf("%w", errRRSigExpired)
	}

	err = checkRRSigSignerName(rrSig, keyTagToDNSKeys)
	if err != nil {
		return err
	}

	matchingDNSKeys := matchingDNSKeysForRRSIG(rrSig, keyTagToDNSKeys)
	if len(matchingDNSKeys) == 0 {
		return fmt.Errorf("%w: for signer %s, algorithm %d and key tag %d",
			errRRSigDNSKey, rrSig.SignerName, rrSig.Algorithm, rrSig.KeyTag)
	}

	for _, dnsKey := range matchingDNSKeys {
		err = rrSig.Verify(dnsKey, rrSet)
		if err == nil {
			return nil
		}
	}

	return err
}

const (
	maxRRSIGValidationsPerRRSet   = 16
	maxRRSIGValidationsPerMessage = 64
)

var (
	errRRSIGValidationRRSetBudgetExceeded = errors.New("RRSIG validation RRSet budget exceeded")
	errRRSIGValidationBudgetExceeded      = errors.New("RRSIG validation message budget exceeded")
)

type rrsigValidationBudget struct {
	remaining uint
}

func newRRSIGValidationBudget() *rrsigValidationBudget {
	return &rrsigValidationBudget{remaining: maxRRSIGValidationsPerMessage}
}

func (b *rrsigValidationBudget) consume() error {
	if b.remaining == 0 {
		return fmt.Errorf("%w: %d", errRRSIGValidationBudgetExceeded, maxRRSIGValidationsPerMessage)
	}
	b.remaining--
	return nil
}

// sortRRSIGsByAlgo sorts RRSIGs by algorithm preference.
func sortRRSIGsByAlgo(rrSigs []*dns.RRSIG) {
	sort.Slice(rrSigs, func(i, j int) bool {
		return lessDNSKeyAlgorithm(rrSigs[i].Algorithm, rrSigs[j].Algorithm)
	})
}

func checkRRSigSignerName(rrSig *dns.RRSIG, keyTagToDNSKeys dnsKeysByTag) error {
	dnsKeys := keyTagToDNSKeys[rrSig.KeyTag]
	if len(dnsKeys) == 0 {
		return fmt.Errorf("%w: for key tag %d", errRRSigDNSKey, rrSig.KeyTag)
	}

	validSignerNames := make([]string, 0, len(dnsKeys))
	for _, dnsKey := range dnsKeys {
		signerName := dnsKey.Header().Name
		if strings.EqualFold(signerName, rrSig.SignerName) {
			return nil
		}

		duplicate := false
		for _, existingSignerName := range validSignerNames {
			if strings.EqualFold(existingSignerName, signerName) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			validSignerNames = append(validSignerNames, signerName)
		}
	}

	quoteStrings(validSignerNames)
	return fmt.Errorf("for %s: %w: %q should be %s",
		rrSigToOwnerTypeCovered(rrSig), errRRSigSignerName,
		rrSig.SignerName, orStrings(validSignerNames))
}
