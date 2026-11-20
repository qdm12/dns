package dnsmsg

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func Test_SetReply(t *testing.T) {
	t.Parallel()

	const requestID uint16 = 1234

	testCases := map[string]struct {
		rcode int
	}{
		"success":         {rcode: dns.RcodeSuccess},
		"name_error":      {rcode: dns.RcodeNameError},
		"server_failure":  {rcode: dns.RcodeServerFailure},
		"refused":         {rcode: dns.RcodeRefused},
		"not_implemented": {rcode: dns.RcodeNotImplemented},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Id:               requestID,
					RecursionDesired: true,
				},
				Question: []dns.Question{{
					Name:   "test.",
					Qtype:  dns.TypeA,
					Qclass: dns.ClassINET,
				}},
			}
			response := &dns.Msg{MsgHdr: dns.MsgHdr{Rcode: testCase.rcode}}

			SetReply(response, request)

			expectedResponse := &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Id:               requestID,
					Response:         true,
					Opcode:           dns.OpcodeQuery,
					RecursionDesired: true,
					// Kept from the response given, whatever it is.
					Rcode: testCase.rcode,
				},
				Question: []dns.Question{{
					Name:   "test.",
					Qtype:  dns.TypeA,
					Qclass: dns.ClassINET,
				}},
			}
			assert.Equal(t, expectedResponse, response)
		})
	}
}
