package dnssec

import (
	"errors"
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

// buildDelegationChain queries the RRs required for the zone validation.
// It begins the queries at the root zone and then go down the delegation
// chain until it reaches the desired zone, or an unsigned zone.
// It returns a delegation chain of signed zones where the
// first signed zone (index 0) is the root zone and the last signed
// zone is the last signed zone, which can be the desired zone.
func buildDelegationChain(handler dns.Handler, desiredZone string, qClass uint16) (
	delegationChain []signedData, err error) {
	zoneNames := desiredZoneToZoneNames(desiredZone)
	delegationChain = make([]signedData, 0, len(zoneNames))

	for _, zoneName := range zoneNames {
		// zoneName iterates in this order: ., com., example.com.
		data, signed, err := queryDelegation(handler, zoneName, qClass)
		if err != nil {
			return nil, fmt.Errorf("querying delegation for desired zone %s: %w",
				desiredZone, err)
		}
		delegationChain = append(delegationChain, data)
		if !signed {
			// first zone without a DS RRSet, but it should
			// have at least one NSEC or NSEC3 RRSet, even for
			// NXDOMAIN responses.
			break
		}
	}

	return delegationChain, nil
}

func desiredZoneToZoneNames(desiredZone string) (zoneNames []string) {
	if desiredZone == "." {
		return []string{"."}
	}

	zoneParts := strings.Split(desiredZone, ".")
	zoneNames = make([]string, len(zoneParts))
	for i := range zoneParts {
		zoneNames[i] = dns.Fqdn(strings.Join(zoneParts[len(zoneParts)-1-i:], "."))
	}
	return zoneNames
}

// queryDelegation obtains the DS RRSet and the DNSKEY RRSet
// for a given zone and class, and creates a signed zone with
// this information. It does not query the (non existent)
// DS record for the root zone, which is the trust root anchor.
func queryDelegation(handler dns.Handler, zone string, qClass uint16) (
	data signedData, signed bool, err error) {
	data.zone = zone
	data.class = qClass

	// TODO set root zone DS here!

	// do not query DS for root zone since its DS record
	// is the trust root anchor.
	if zone != "." {
		data.dsResponse, err = queryDS(handler, zone, qClass)
		if err != nil {
			return signedData{}, false, fmt.Errorf("querying DS record: %w", err)
		}

		if data.dsResponse.isNoData() || data.dsResponse.isNXDomain() {
			// If no DS RRSet is found, the entire zone is unsigned.
			// This also means no DNSKEY RRSet exists, since child zones are
			// also unsigned, so return with the error errZoneHasNoDSRcord
			// to signal the caller to stop the delegation chain queries for
			// child zones when encountering a zone with no DS RRSet.
			return data, false, nil
		}
	}

	data.dnsKeyResponse, err = queryDNSKeys(handler, zone, qClass)
	if err != nil {
		return signedData{}, true, fmt.Errorf("querying DNSKEY record: %w", err)
	}

	return data, true, nil
}

var (
	ErrDSAndNSECAbsent = errors.New("zone has no DS record and no NSEC record")
)

func queryDS(handler dns.Handler, zone string, qClass uint16) (
	response dnssecResponse, err error) {
	response, err = queryRRSets(handler, zone, qClass, dns.TypeDS)
	switch {
	case err != nil:
		return dnssecResponse{}, err
	case !response.isSigned():
		// no signed DS answer and no NSEC/NSEC3 authority RR
		return dnssecResponse{}, wrapError(
			zone, qClass, dns.TypeDS, ErrDSAndNSECAbsent)
	case response.isNXDomain(), response.isNoData():
		// there is one or more NSEC/NSEC3 authority RRSets.
		return response, nil
	}
	// signed answer RRSet(s)

	// Double check we only have 1 DS RRSet.
	// TODO remove?
	err = dnssecRRSetsIsSingleOfType(response.answerRRSets, dns.TypeDS)
	if err != nil {
		return dnssecResponse{},
			wrapError(zone, qClass, dns.TypeDS, err)
	}

	return response, nil
}

// queryDNSKeys queries the DNSKEY records for a given signed zone
// containing a DS RRSet. It returns an error if the DNSKEY RRSet is
// missing or is unsigned.
// Note this returns all the DNSKey RRs, even non-zone ones.
func queryDNSKeys(handler dns.Handler, qname string, qClass uint16) (
	response dnssecResponse, err error) {
	// DNSKey RRSet(s) should be present so the NSEC/NSEC3 RRSet is ignored.
	response, err = queryRRSets(handler, qname, qClass, dns.TypeDNSKEY)
	switch {
	case err != nil:
		return dnssecResponse{}, err
	case !response.isSigned(), response.isNoData(): // cannot be NXDOMAIN
		// no signed DNSKEY answer
		return dnssecResponse{}, fmt.Errorf("for %s: %w",
			nameClassTypeToString(qname, qClass, dns.TypeDNSKEY),
			ErrDNSKeyNotFound)
	}

	// Double check we only have 1 DNSKEY RRSet.
	// TODO remove?
	err = dnssecRRSetsIsSingleOfType(response.answerRRSets, dns.TypeDNSKEY)
	if err != nil {
		return dnssecResponse{},
			wrapError(qname, qClass, dns.TypeDNSKEY, err)
	}

	return response, nil
}
