package console

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func Test_Formatter_Request(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request *dns.Msg
		s       string
	}{
		"no question": {
			request: new(dns.Msg),
			s:       "id: 0; no question",
		},
		"single question": {
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "github.com.", Qclass: dns.ClassINET, Qtype: dns.TypeA},
				},
			},
			s: "id: 0; question: github.com. IN A",
		},
		"two questions": {
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "github.com.", Qclass: dns.ClassINET, Qtype: dns.TypeA},
					{Name: "github.com.", Qclass: dns.ClassINET, Qtype: dns.TypeAAAA},
				},
			},
			s: "id: 0; questions: github.com. IN A, github.com. IN AAAA",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			formatter := new(Formatter)
			s := formatter.Request(testCase.request)

			assert.Equal(t, testCase.s, s)
		})
	}
}
