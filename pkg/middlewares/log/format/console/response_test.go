package console

import (
	"net"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func Test_Formatter_Response(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		response *dns.Msg
		s        string
	}{
		"no answer": {
			response: new(dns.Msg),
			s:        "id: 0; no answer",
		},
		"single answer": {
			response: &dns.Msg{
				Answer: []dns.RR{
					&dns.A{A: net.IP{1, 2, 3, 4}},
				},
			},
			s: "id: 0; answer: 0 CLASS0 None 1.2.3.4",
		},
		"two answers": {
			response: &dns.Msg{
				Answer: []dns.RR{
					&dns.A{A: net.IP{1, 2, 3, 4}},
					&dns.AAAA{AAAA: net.IPv6loopback},
				},
			},
			s: `id: 0; answers: [
  0 CLASS0 None 1.2.3.4,
  0 CLASS0 None ::1
]`,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			formatter := new(Formatter)
			s := formatter.Response(testCase.response)

			assert.Equal(t, testCase.s, s)
		})
	}
}
