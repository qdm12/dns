//go:build integration
// +build integration

package dnssec

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/stateful"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Validator_Validate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request              *dns.Msg
		errWrapped           error
		errMessage           string
		assertVolatileAnswer bool
	}{
		"exists_not_signed": {
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "test.github.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
				},
			},
		},
		"exists_signed": {
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "icann.org.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
				},
			},
		},
		"nodata_nsec3": {
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "icann.org.", Qtype: dns.TypeMD, Qclass: dns.ClassINET},
				},
			},
		},
		"nxdomain_nsec3": {
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "xyz.icann.org.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
				},
			},
		},
		"cname_chain_signed_unsigned_dns1.nextdns.io": {
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "dns1.nextdns.io.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
				},
			},
		},
		"cname_chain_signed_unsigned_acme-v02.api.letsencrypt.org_A": {
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "acme-v02.api.letsencrypt.org.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
				},
			},
		},
		"cname_chain_signed_unsigned_acme-v02.api.letsencrypt.org_AAAA": {
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "acme-v02.api.letsencrypt.org.", Qtype: dns.TypeAAAA, Qclass: dns.ClassINET},
				},
			},
		},
		"cname_chain_signed_unsigned_mrodevicemgr.officeapps.live.com_A": {
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "mrodevicemgr.officeapps.live.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
				},
			},
			// Endpoint A records rotate quickly, so compare only stable properties.
			assertVolatileAnswer: true,
		},
		"nxdomain_nsec": {
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "x.cloudflare.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
				},
			},
		},
		"reverse_ptr_nxdomain_44.10.10.10.in-addr.arpa": {
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "44.10.10.10.in-addr.arpa.", Qtype: dns.TypePTR, Qclass: dns.ClassINET},
				},
			},
		},
		"reverse_ptr_nodata_209.85.244.167.in-addr.arpa": {
			// This reverse zone returns signed NODATA for DS queries with NSEC
			// interval coverage proofs (not exact name match), testing that our
			// NSEC validation correctly handles interval-based coverage.
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "209.85.244.167.in-addr.arpa.", Qtype: dns.TypePTR, Qclass: dns.ClassINET},
				},
			},
		},
		"a_and_cname": {
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "sigok.ippacket.stream.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
				},
			},
		},
		//
		// Special cases
		//
		"dnssec_failed_by_upstream": {
			// One can also try rhybar.cz.
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "dnssec-failed.org.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
				},
			},
			errWrapped: errRcodeBad,
			errMessage: "running desired query: " +
				"for dnssec-failed.org. IN A: " +
				"bad response rcode: SERVFAIL",
		},
		"signed_answer_insecure_parent": {
			// The answer is a NODATA with an NSEC RRSet signed by whispersystems.org.
			// The parent zone whispersystems.org. has DNSKEYs (ZSK+KSK) but
			// no DS record, so it is therefore insecure and so is the answer.
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "textsecure-service.whispersystems.org.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
				},
			},
		},
		"signed_answer_insecure_parent_contentproxy.signal.org": {
			// contentproxy.signal.org. is delegated from signal.org. without DS.
			// Validation should treat it as insecure delegation and not fail
			// delegation chain construction.
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "contentproxy.signal.org.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
				},
			},
		},
		"signed_answer_insecure_parent_deploy.mopinion.com": {
			// deploy.mopinion.com. is in an unsigned delegation path and should be
			// treated as insecure rather than bogus even when DS denial proofs
			// are not present from a resolver.
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "deploy.mopinion.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
				},
			},
			assertVolatileAnswer: true,
		},
		"nxdomain_2_rrsigs_per_nsec": {
			// There are two RRSIGs per NSEC RR, each with a
			// different algorithm. This is to allow transitioning
			// from one weaker/older algorithm to a stronger/newer one.
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "xyzzy14.sdsmt.edu.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
				},
			},
		},
		"nodata_2_rrsigs_dnskey": {
			// The DNSKEY RRSet of vip.icann.org. is signed by two RRSIGs,
			// one validating against the ZSK of icann.org. and the other
			// validating against the KSK of icann.org. This is valid although
			// not very conventional.
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "vip.icann.org.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
				},
			},
		},
		"wildcard_expanded": {
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "b.zahrarestaurant.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
				},
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testCase.request.RecursionDesired = true
			testCase.request.Id = dns.Id()
			requestCopy := testCase.request.Copy()

			handler := newIntegTestHandler(t)
			validator := New()

			response, err := validator.Validate(testCase.request, handler)

			require.ErrorIs(t, err, testCase.errWrapped)

			if testCase.assertVolatileAnswer {
				require.Nil(t, err)
				require.NotNil(t, response)
				assert.Equal(t, dns.RcodeSuccess, response.Rcode)
				assert.NotEmpty(t, response.Answer)

				var hasA bool
				for _, answer := range response.Answer {
					if answer.Header().Rrtype == dns.TypeA {
						hasA = true
						break
					}
				}
				assert.True(t, hasA)
				return
			}

			var expectedResponse *dns.Msg
			if testCase.errWrapped != nil {
				assert.EqualError(t, err, testCase.errMessage)
			} else { // no error, fetch expected response
				statefulWriter := stateful.NewWriter()
				requestCopy.Id = dns.Id()
				handler.ServeDNS(statefulWriter, requestCopy)
				expectedResponse = statefulWriter.Response
				// DNSSEC does not do recursion for now
				expectedResponse.RecursionAvailable = false
			}

			assertResponsesEqual(t, expectedResponse, response)
		})
	}
}

type integTestHandler struct {
	t      *testing.T
	client *dns.Client
	dialer *net.Dialer
}

func newIntegTestHandler(t *testing.T) *integTestHandler {
	return &integTestHandler{
		t:      t,
		client: &dns.Client{},
		dialer: &net.Dialer{},
	}
}

func (h *integTestHandler) ServeDNS(w dns.ResponseWriter, request *dns.Msg) {
	request = request.Copy()

	deadline, ok := h.t.Deadline()
	if !ok {
		deadline = time.Now().Add(4 * time.Second)
	}
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	const maxTries = 3
	success := false
	var response *dns.Msg
	for i := range maxTries {
		const timeout = time.Second
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Try a new UDP connection on each try
		netConn, err := h.dialer.DialContext(ctx, "udp", "1.1.1.1:53")
		require.NoError(h.t, err)
		dnsConn := &dns.Conn{Conn: netConn}

		response, _, err = h.client.ExchangeWithConnContext(ctx, request, dnsConn)
		if err != nil {
			_ = dnsConn.Close()
			h.t.Logf("try %d of %d: %s", i+1, maxTries, err)
			continue
		}

		err = dnsConn.Close()
		require.NoError(h.t, err)

		success = true
		break
	}

	if !success {
		h.t.Fatalf("could not communicate with DNS server after %d tries", maxTries)
	}

	if !response.Truncated {
		// Remove TTL fields from rrset
		for i := range response.Answer {
			response.Answer[i].Header().Ttl = 0
		}

		_ = w.WriteMsg(response)
		return
	}

	// Retry with TCP
	netConn, err := h.dialer.DialContext(ctx, "tcp", "1.1.1.1:53")
	require.NoError(h.t, err)

	dnsConn := &dns.Conn{Conn: netConn}
	response, _, err = h.client.ExchangeWithConnContext(ctx, request, dnsConn)
	require.NoError(h.t, err)

	err = dnsConn.Close()
	require.NoError(h.t, err)

	_ = w.WriteMsg(response)
}

func assertResponsesEqual(t *testing.T, a, b *dns.Msg) {
	if a == nil {
		require.Nil(t, b)
		return
	}
	require.NotNil(t, b)

	// Remove TTL fields from answer and authority
	for i := range a.Answer {
		a.Answer[i].Header().Ttl = 0
	}
	for i := range a.Ns {
		a.Ns[i].Header().Ttl = 0
	}
	for i := range b.Answer {
		b.Answer[i].Header().Ttl = 0
	}
	for i := range b.Ns {
		b.Ns[i].Header().Ttl = 0
	}

	a.Id = 0
	b.Id = 0

	assert.Equal(t, a, b)
}
