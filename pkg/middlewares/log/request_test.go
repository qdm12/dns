package log

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func Test_requestString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request *dns.Msg
		s       string
	}{
		"no question": {
			request: new(dns.Msg),
			s:       "[empty request]",
		},
		"single question": {
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "github.com.", Qclass: dns.ClassINET, Qtype: dns.TypeA},
				},
			},
			s: "github.com. IN A",
		},
		"two questions": {
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "github.com.", Qclass: dns.ClassINET, Qtype: dns.TypeA},
					{Name: "ignored.com.", Qclass: dns.ClassINET, Qtype: dns.TypeA},
				},
			},
			s: "github.com. IN A",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := requestString(testCase.request)

			assert.Equal(t, testCase.s, s)
		})
	}
}
