package dnssec

import (
	"errors"
	"fmt"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/stateful"
)

var errRcodeBad = errors.New("bad response rcode")

func queryRRSets(handler dns.Handler, zone string,
	qClass, qType uint16,
) (response dnssecResponse, err error) {
	request := newEDNSRequest(zone, qClass, qType)

	statefulWriter := stateful.NewWriter()
	handler.ServeDNS(statefulWriter, request)
	dnsResponse := statefulWriter.Response
	response.rcode = dnsResponse.Rcode

	switch {
	case dnsResponse.Rcode == dns.RcodeSuccess && len(dnsResponse.Answer) > 0:
		// Success and we have at least one answer RR.
		response.answerRRSets, err = groupRRs(dnsResponse.Answer)
		if err != nil {
			return dnssecResponse{}, fmt.Errorf(
				"grouping answer RRSets for %s: %w",
				nameClassTypeToString(zone, qClass, qType), err)
		}

		if !response.isSigned() {
			// We have all unsigned answers
			return response, nil
		}

		// Every RRSet has at least one RRSIG associated with it.
		// The caller should then verify the RRSIGs and MAY need
		// NSEC or NSEC3 RRSets from the authority section to verify
		// it does not match a wildcard.
		response.authorityRRSets, err = groupRRs(dnsResponse.Ns)
		if err != nil {
			return dnssecResponse{}, fmt.Errorf(
				"grouping authority RRSets for %s: %w",
				nameClassTypeToString(zone, qClass, qType), err)
		}

		return response, nil
	case dnsResponse.Rcode == dns.RcodeSuccess && len(dnsResponse.Answer) == 0,
		dnsResponse.Rcode == dns.RcodeNameError:
		// NXDOMAIN or NODATA response, we need to verify the negative
		// response with the query authority section NSEC/NSEC3 RRSet
		// or verify the zone is insecure.
		// If the zone is insecure, the caller verifies the zone is
		// insecure using the NSEC/NSEC3 records of the authority
		// section of the DS query for that zone, or any first of
		// its parent zone with an NSEC/NSEC3 record for that zone,
		// walking towards the root zone.
		// There is no difference in handling if we received a NODATA
		// or NXDOMAIN response.

		if len(dnsResponse.Ns) == 0 {
			// No authority RR so there cannot be any NSEC/NSEC3 RRSet,
			// the zone is thus insecure.
			return response, nil
		}

		response.authorityRRSets, err = groupRRs(dnsResponse.Ns)
		if err != nil {
			return dnssecResponse{}, fmt.Errorf(
				"grouping authority RRSets for %s: %w",
				nameClassTypeToString(zone, qClass, qType), err)
		}

		// Negative responses must have signed NSEC/NSEC3 RRSets for validity.
		// RFC 4034 §4.1 and RFC 5155 require NSEC/NSEC3 to be signed when
		// used in negative assertions. Unsigned NSEC/NSEC3 cannot prove non-existence.
		err = validateNegativeResponseProofSignatures(response.authorityRRSets)
		if err != nil {
			return dnssecResponse{}, fmt.Errorf(
				"validating negative response proof signatures for %s: %w",
				nameClassTypeToString(zone, qClass, qType), err)
		}

		return response, nil
	default: // other error
		// If the response Rcode is dns.RcodeServerFailure,
		// this may mean DNSSEC validation failed on the upstream server.
		// https://www.ietf.org/rfc/rfc4033.txt
		// This specification only defines how security-aware name servers can
		// signal non-validating stub resolvers that data was found to be bogus
		// (using RCODE=2, "Server Failure"; see [RFC4035]).
		return dnssecResponse{}, fmt.Errorf(
			"for %s: %w: %s",
			nameClassTypeToString(zone, qClass, qType),
			errRcodeBad, dns.RcodeToString[dnsResponse.Rcode])
	}
}

var (
	errRRSetSignedAndUnsigned        = errors.New("mix of signed and unsigned RRSets")
	errRRSigForNoRRSet               = errors.New("RRSIG for no RRSet")
	errNegativeResponseUnsignedNSEC  = errors.New("negative response has unsigned NSEC")
	errNegativeResponseUnsignedNSEC3 = errors.New("negative response has unsigned NSEC3")
)

// validateNegativeResponseProofSignatures ensures NSEC/NSEC3 RRSets
// used in negative proofs are all signed. Per RFC 4034 and RFC 5155,
// unsigned NSEC/NSEC3 cannot serve as proof of non-existence.
func validateNegativeResponseProofSignatures(rrSets []dnssecRRSet) (err error) {
	for _, rrSet := range rrSets {
		if len(rrSet.rrSigs) != 0 {
			continue
		}
		switch rrSet.qtype() {
		case dns.TypeNSEC:
			err = errNegativeResponseUnsignedNSEC
		case dns.TypeNSEC3:
			err = errNegativeResponseUnsignedNSEC3
		default:
			continue
		}
		owner := rrSet.rrSet[0].Header().Name
		return fmt.Errorf("%w: RRSet owner %s has no RRSIG", err, owner)
	}
	return nil
}

// groupRRs groups RRs by type AND owner AND class, returning a slice
// of 'DNSSEC RRSets' where each contains at least one RR,
// and zero or one RRSIG signature.
// Regarding the RRSig validity requirements listed in
// https://datatracker.ietf.org/doc/html/rfc4035#section-5.3.1
//
// The following requirements are fullfiled by design:
//   - The RRSIG RR and the RRset MUST have the same owner name and the
//     same class
//   - The RRSIG RR's Type Covered field MUST equal the RRset's type.
//
// And the function returns an error for the following unmet requirement:
//   - The number of labels in the RRset owner name MUST be greater than
//     or equal to the value in the RRSIG RR's Labels field.
//
// The following requirements are enforced at a later stage:
//   - The RRSIG RR's Signer's Name field MUST be the name of the zone
//     that contains the RRset.
//   - The validator's notion of the current time MUST be less than or
//     equal to the time listed in the RRSIG RR's Expiration field.
func groupRRs(rrs []dns.RR) (dnssecRRSets []dnssecRRSet, err error) {
	maxRRSets := len(rrs) // all unsigned RRs
	dnssecRRSets = make([]dnssecRRSet, 0, maxRRSets)
	type typeZoneKey struct {
		rrType uint16
		owner  string
		class  uint16
	}
	typeZoneToIndex := make(map[typeZoneKey]int, maxRRSets)

	for _, rr := range rrs {
		header := rr.Header()
		typeZoneKey := typeZoneKey{
			owner: header.Name,
			class: header.Class,
		}

		rrType := header.Rrtype
		if rrType == dns.TypeRRSIG {
			rrsig := mustRRToRRSig(rr)
			err = rrsigInitialChecks(rrsig)
			if err != nil {
				return nil, err
			}

			typeZoneKey.rrType = rrsig.TypeCovered
			i, ok := typeZoneToIndex[typeZoneKey]
			if !ok {
				dnssecRRSets = append(dnssecRRSets, dnssecRRSet{})
				i = len(dnssecRRSets) - 1
				typeZoneToIndex[typeZoneKey] = i
			}

			dnssecRRSets[i].rrSigs = append(dnssecRRSets[i].rrSigs, rrsig)
			continue
		}

		typeZoneKey.rrType = rrType
		i, ok := typeZoneToIndex[typeZoneKey]
		if !ok {
			dnssecRRSets = append(dnssecRRSets, dnssecRRSet{})
			i = len(dnssecRRSets) - 1
			typeZoneToIndex[typeZoneKey] = i
		}
		dnssecRRSets[i].rrSet = append(dnssecRRSets[i].rrSet, rr)
	}

	err = validateOwnerSigningConsistency(dnssecRRSets)
	if err != nil {
		return nil, err
	}

	// Verify built DNSSEC RRSets are well formed.
	for _, dnssecRRSet := range dnssecRRSets {
		if len(dnssecRRSet.rrSigs) > 0 && len(dnssecRRSet.rrSet) == 0 {
			return nil, fmt.Errorf("for RRSet %s %s: %w",
				dnssecRRSet.rrSigs[0].Hdr.Name,
				dns.TypeToString[dnssecRRSet.rrSigs[0].TypeCovered],
				errRRSigForNoRRSet)
		}
	}

	return dnssecRRSets, nil
}

// validateOwnerSigningConsistency checks each owner name has only signed
// or only unsigned RRSets. Different owner names (from different zones in a
// CNAME chain) may legitimately have different signing status.
func validateOwnerSigningConsistency(rrSets []dnssecRRSet) error {
	type ownerStats struct{ signed, unsigned int }
	ownerSigningMap := make(map[string]*ownerStats, len(rrSets))
	for _, rrSet := range rrSets {
		var owner string
		if len(rrSet.rrSet) > 0 {
			owner = rrSet.rrSet[0].Header().Name
		} else {
			// RRSIG-only RRSet; errRRSigForNoRRSet will be returned by groupRRs.
			owner = rrSet.rrSigs[0].Hdr.Name
		}

		stats := ownerSigningMap[owner]
		if stats == nil {
			stats = &ownerStats{}
			ownerSigningMap[owner] = stats
		}
		if len(rrSet.rrSigs) > 0 {
			stats.signed++
		} else {
			stats.unsigned++
		}
	}

	for owner, stats := range ownerSigningMap {
		if stats.signed > 0 && stats.unsigned > 0 {
			return fmt.Errorf("%w: owner %s has %d signed and %d unsigned RRSets",
				errRRSetSignedAndUnsigned, owner, stats.signed, stats.unsigned)
		}
	}

	return nil
}
