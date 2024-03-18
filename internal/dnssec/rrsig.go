package dnssec

import (
	"errors"
	"fmt"
	"sort"
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
	ownerLabelsCount := uint8(dns.CountLabel(rrSig.Hdr.Name))
	return rrSig.Labels < ownerLabelsCount
}

var (
	ErrRRSigLabels = errors.New("RRSIG labels greater than owner labels")
)

// See https://datatracker.ietf.org/doc/html/rfc4035#section-5.3.1
func rrsigInitialChecks(rrsig *dns.RRSIG) (err error) {
	rrSetOwner := rrsig.Hdr.Name

	err = rrSigCheckSignerName(rrsig)
	if err != nil {
		return err
	}

	if int(rrsig.Labels) > dns.CountLabel(rrSetOwner) {
		// The number of labels in the RRset owner name MUST be greater than
		// or equal to the value in the RRSIG RR's Labels field.
		return fmt.Errorf("for %s: %w: RRSig labels field is %d and owner is %d labels",
			rrSigToOwnerTypeCovered(rrsig), ErrRRSigLabels,
			rrsig.Labels, dns.CountLabel(rrSetOwner))
	}

	return nil
}

func verifyRRSetsRRSig(answerRRSets []dnssecRRSet, keyTagToDNSKey map[uint16]*dns.DNSKEY) (err error) {
	for _, signedRRSet := range answerRRSets {
		err = verifyRRSetRRSigs(signedRRSet.rrSet,
			signedRRSet.rrSigs, keyTagToDNSKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func verifyRRSetRRSigs(rrSet []dns.RR, rrSigs []*dns.RRSIG,
	keyTagToDNSKey map[uint16]*dns.DNSKEY) (
	err error) {
	if len(rrSet) == 0 || len(rrSigs) == 0 {
		panic("no rrs or rrsigs")
	}

	if len(rrSigs) == 1 {
		return verifyRRSetRRSig(rrSet, rrSigs[0], keyTagToDNSKey)
	}

	// Multiple RRSIGs for the same RRSet, sort them by algorithm preference
	// and try each one until one succeeds. This is rather undocumented,
	// but one signature verified should be enough to validate the RRSet,
	// even if other signatures fail to verify successfully.
	sortRRSIGsByAlgo(rrSigs)

	errs := new(joinedErrors)
	for _, rrSig := range rrSigs {
		if !rrSig.ValidityPeriod(time.Now()) {
			errs.add(fmt.Errorf("%w", ErrRRSigExpired))
			continue
		}

		keyTag := rrSig.KeyTag
		dnsKey, ok := keyTagToDNSKey[keyTag]
		if !ok {
			errs.add(fmt.Errorf("%w: in %d DNSKEY(s) for key tag %d",
				ErrRRSigDNSKeyTag, len(keyTagToDNSKey), keyTag))
			continue
		}

		err = rrSig.Verify(dnsKey, rrSet)
		if err != nil {
			errs.add(err)
			continue
		}

		return nil
	}

	return fmt.Errorf("%d RRSIGs failed to validate the RRSet: %w",
		len(rrSigs), errs)
}

var (
	ErrRRSigDNSKeyTag = errors.New("DNSKEY not found")
	ErrRRSigExpired   = errors.New("RRSIG has expired")
)

func verifyRRSetRRSig(rrSet []dns.RR, rrSig *dns.RRSIG,
	keyTagToDNSKey map[uint16]*dns.DNSKEY) (err error) {
	if !rrSig.ValidityPeriod(time.Now()) {
		return fmt.Errorf("%w", ErrRRSigExpired)
	}

	keyTag := rrSig.KeyTag
	dnsKey, ok := keyTagToDNSKey[keyTag]
	if !ok {
		return fmt.Errorf("%w: in %d DNSKEY(s) for key tag %d",
			ErrRRSigDNSKeyTag, len(keyTagToDNSKey), keyTag)
	}

	err = rrSig.Verify(dnsKey, rrSet)
	if err != nil {
		return err
	}

	return nil
}

// sortRRSIGsByAlgo sorts RRSIGs by algorithm preference.
func sortRRSIGsByAlgo(rrSigs []*dns.RRSIG) {
	sort.Slice(rrSigs, func(i, j int) bool {
		return lessDNSKeyAlgorithm(rrSigs[i].Algorithm, rrSigs[j].Algorithm)
	})
}

var (
	ErrRRSigSignerName = errors.New("signer name is not valid")
)

// The RRSIG RR's Signer's Name field MUST be the
// name of the zone that contains the RRset.
func rrSigCheckSignerName(rrSig *dns.RRSIG) (err error) {
	var validSignerNames []string
	switch rrSig.TypeCovered {
	case dns.TypeDS, dns.TypeCNAME, dns.TypeNSEC3:
		validSignerNames = []string{parentName(rrSig.Hdr.Name)}
	default:
		// For NSEC RRs, the signer name must be the apex name which
		// can be the owner or the parent of the owner of the RRSIG.
		// For example:
		// p.example.com. 3601 IN NSEC   r.example.com. A RRSIG NSEC
		// p.example.com. 3601 IN RRSIG  NSEC 13 3 3601 20240111000000 20231221000000 42950 example.com. 0se..m GY..w==
		// example.com.   3601 IN NSEC   l.example.com. A NS SOA RRSIG NSEC DNSKEY
		// example.com.   3601 IN RRSIG  NSEC 13 2 3601 20240111000000 20231221000000 42950 example.com. pe..B 4V..Q==

		// For other RRs, such as A, the signer name must be the owner
		// or the parent of the owner, for example for sigok.ippacket.stream.
		// the A record RRSIG owner is sigok.rsa2048-sha256.ippacket.stream.
		// and signer name is rsa2048-sha256.ippacket.stream.
		validSignerNames = []string{rrSig.Hdr.Name, parentName(rrSig.Hdr.Name)}
	}

	if isOneOf(rrSig.SignerName, validSignerNames...) {
		return nil
	}

	quoteStrings(validSignerNames)
	return fmt.Errorf("for %s: %w: %q should be %s",
		rrSigToOwnerTypeCovered(rrSig), ErrRRSigSignerName,
		rrSig.SignerName, orStrings(validSignerNames))
}

func parentName(name string) (parent string) {
	const offset = 0
	nextLabelStart, end := dns.NextLabel(name, offset)
	if end {
		// parent of 'tld.' is '.' and parent of '.' is '.'
		return "."
	}
	return name[nextLabelStart:]
}
