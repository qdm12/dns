package lru

import (
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

type entry struct {
	key      string // from the DNS request
	expUnix  int64  // from the DNS response
	response *dns.Msg
}

func makeKey(request *dns.Msg) (key string) {
	question := request.Question[0]
	key = strings.ToLower(question.Name) + "|" +
		strconv.FormatUint(uint64(question.Qtype), 10) + "|" +
		strconv.FormatUint(uint64(question.Qclass), 10)
	return key
}

// getExpUnix computes the expiration time for a cached DNS response,
// considering both standard TTLs and DNSSEC-specific constraints.
// For DNSSEC-signed responses, it also considers:
// - RRSIG Expiration time (absolute Unix timestamp).
// - NSEC/NSEC3 TTL in authority section.
// - SOA MINIMUM field for negative responses.
func getExpUnix(response *dns.Msg, nowUnix int64) (expUnix int64) {
	secondsLeft := ^uint32(0)
	for _, rr := range response.Answer {
		// For RRSIG records, use the Expiration field, not the TTL
		if rrsig, ok := rr.(*dns.RRSIG); ok {
			var secondsUntilExpiration uint32
			if nowUnix < int64(rrsig.Expiration) {
				secondsUntilExpiration = rrsig.Expiration - uint32(nowUnix) //nolint:gosec
			}
			secondsLeft = min(secondsLeft, secondsUntilExpiration)
			continue // Skip the Ttl check for RRSIG
		}

		ttl := rr.Header().Ttl
		if ttl < secondsLeft {
			secondsLeft = ttl
		}
	}

	// Check Authority section for NSEC/NSEC3 and SOA (for negative responses or proofs)
	for _, rr := range response.Ns {
		if rrsig, ok := rr.(*dns.RRSIG); ok {
			var secondsUntilExpiration uint32
			if nowUnix < int64(rrsig.Expiration) {
				secondsUntilExpiration = rrsig.Expiration - uint32(nowUnix) //nolint:gosec
			}
			secondsLeft = min(secondsLeft, secondsUntilExpiration)
			continue // Skip the Ttl check for RRSIG
		}

		ttl := rr.Header().Ttl
		secondsLeft = min(secondsLeft, ttl)

		// For negative responses, check SOA MINIMUM field
		if soa, ok := rr.(*dns.SOA); ok {
			secondsLeft = min(secondsLeft, soa.Minttl)
		}
	}

	// Handle max value case (no records found)
	if secondsLeft == ^uint32(0) {
		secondsLeft = 0
	}

	return nowUnix + int64(secondsLeft)
}

// verifyRRSIGValidity checks if any RRSIGs in the response have expired.
// This is called during cache retrieval to ensure cached DNSSEC responses
// are still valid at the time of use.
func verifyRRSIGValidity(response *dns.Msg, nowUnix int64) bool {
	nowUnixUint32 := uint32(nowUnix) //nolint:gosec

	for _, rr := range response.Answer {
		if rrsig, ok := rr.(*dns.RRSIG); ok {
			if nowUnixUint32 >= rrsig.Expiration {
				return false
			}
		}
	}

	for _, rr := range response.Ns {
		if rrsig, ok := rr.(*dns.RRSIG); ok {
			if nowUnixUint32 >= rrsig.Expiration {
				return false
			}
		}
	}

	return true
}
