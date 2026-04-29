package dnssec

import (
	"errors"
	"fmt"
	"sort"

	"github.com/miekg/dns"
)

type trustAnchorSet struct {
	dsRecords []dns.DS
}

func newTrustAnchorSet(dsRecords []dns.DS) *trustAnchorSet {
	set := &trustAnchorSet{
		dsRecords: deduplicateTrustAnchors(dsRecords),
	}
	return set
}

func (s *trustAnchorSet) cloneDSRecords() []dns.DS {
	cloned := make([]dns.DS, len(s.dsRecords))
	copy(cloned, s.dsRecords)
	return cloned
}

func (v *Validator) setRootTrustAnchors(dsRecords []dns.DS) {
	v.rootTrustAnchors.Store(newTrustAnchorSet(dsRecords))
}

func (v *Validator) rootTrustAnchorsDSRecords() []dns.DS {
	set := v.rootTrustAnchors.Load()
	if set == nil {
		return nil
	}
	return set.cloneDSRecords()
}

func (v *Validator) RefreshRootTrustAnchors(handler dns.Handler) error {
	updatedDSRecords, err := refreshRootTrustAnchors(handler,
		v.rootTrustAnchorsDSRecords())
	if err != nil {
		return err
	}
	v.setRootTrustAnchors(updatedDSRecords)
	return nil
}

const (
	rootKSK2017KeyTag uint16 = 20326
	rootKSK2024KeyTag uint16 = 38696
)

func defaultRootTrustAnchors() []dns.DS {
	return []dns.DS{
		{
			Hdr:        dns.RR_Header{Name: ".", Rrtype: dns.TypeDS, Class: dns.ClassINET},
			Algorithm:  dns.RSASHA256,
			DigestType: dns.SHA256,
			KeyTag:     rootKSK2017KeyTag,
			Digest:     "E06D44B80B8F1D39A95C0B0D7C65D08458E880409BBC683457104237C7F8EC8D",
		},
		{
			Hdr:        dns.RR_Header{Name: ".", Rrtype: dns.TypeDS, Class: dns.ClassINET},
			Algorithm:  dns.RSASHA256,
			DigestType: dns.SHA256,
			KeyTag:     rootKSK2024KeyTag,
			Digest:     "683D2D0ACB8C9B712A1948B27F741219298D0A450D612C483AF444A4C0FB2B16",
		},
	}
}

func refreshRootTrustAnchors(handler dns.Handler, currentDSRecords []dns.DS) (
	updatedDSRecords []dns.DS, err error,
) {
	rootDNSKEYResponse, err := queryDNSKeys(handler, ".", dns.ClassINET)
	if err != nil {
		return nil, fmt.Errorf("querying root DNSKEY RRSet: %w", err)
	}

	rootDNSKEYRRSet := rootDNSKEYResponse.onlyAnswerRRSet()
	rootKeyTagToDNSKeys := makeKeyTagToDNSKeys(rootDNSKEYRRSet)
	err = verifyRRSetRRSigs(rootDNSKEYRRSet, rootDNSKEYResponse.onlyAnswerRRSigs(),
		rootKeyTagToDNSKeys, newRRSIGValidationBudget())
	if err != nil {
		return nil, fmt.Errorf("verifying root DNSKEY RRSet signatures: %w", err)
	}

	err = verifyRootTrustAnchors(currentDSRecords, rootKeyTagToDNSKeys)
	if err != nil {
		return nil, fmt.Errorf(
			"verifying current root trust anchors against root DNSKEY RRSet: %w",
			err,
		)
	}

	updatedDSRecords = deriveTrustAnchorsFromDNSKEYRRSet(rootDNSKEYRRSet)
	if len(updatedDSRecords) == 0 {
		return nil, errors.New("no root trust anchors found")
	}

	return updatedDSRecords, nil
}

func verifyRootTrustAnchors(rootTrustAnchors []dns.DS,
	rootKeyTagToDNSKeys dnsKeysByTag,
) error {
	if len(rootTrustAnchors) == 0 {
		return errors.New("no root trust anchors found")
	}

	errs := new(joinedErrors)
	for _, rootTrustAnchor := range rootTrustAnchors {
		err := verifyDS(&rootTrustAnchor, rootKeyTagToDNSKeys)
		if err == nil {
			return nil
		}
		errs.add(fmt.Errorf("for root anchor key tag %d: %w",
			rootTrustAnchor.KeyTag, err))
	}

	return errs
}

func deriveTrustAnchorsFromDNSKEYRRSet(dnsKeyRRSet []dns.RR) []dns.DS {
	dsRecords := make([]dns.DS, 0, len(dnsKeyRRSet))
	for _, rr := range dnsKeyRRSet {
		dnsKey := mustRRToDNSKey(rr)
		const (
			rootZoneProtocol = 3
			sepFlagMask      = 1
		)
		if dnsKey.Protocol != rootZoneProtocol || dnsKey.Flags&dns.ZONE == 0 ||
			dnsKey.Flags&sepFlagMask == 0 {
			continue
		}
		ds := dnsKey.ToDS(dns.SHA256)
		if ds == nil {
			continue
		}
		dsRecords = append(dsRecords, *ds)
	}
	return deduplicateTrustAnchors(dsRecords)
}

func deduplicateTrustAnchors(dsRecords []dns.DS) []dns.DS {
	unique := make(map[string]dns.DS, len(dsRecords))
	for _, dsRecord := range dsRecords {
		key := fmt.Sprintf("%s|%d|%d|%d|%s", dsRecord.Header().Name,
			dsRecord.KeyTag, dsRecord.Algorithm, dsRecord.DigestType,
			dsRecord.Digest)
		unique[key] = dsRecord
	}

	deduplicated := make([]dns.DS, 0, len(unique))
	for _, dsRecord := range unique {
		deduplicated = append(deduplicated, dsRecord)
	}
	sort.Slice(deduplicated, func(i, j int) bool {
		if deduplicated[i].KeyTag != deduplicated[j].KeyTag {
			return deduplicated[i].KeyTag < deduplicated[j].KeyTag
		}
		if deduplicated[i].DigestType != deduplicated[j].DigestType {
			return deduplicated[i].DigestType < deduplicated[j].DigestType
		}
		return deduplicated[i].Digest < deduplicated[j].Digest
	})
	return deduplicated
}
