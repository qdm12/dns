package stateful

import (
	"net"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func Test_Writer(t *testing.T) {
	t.Parallel()

	dummyResponse := &dns.Msg{Answer: []dns.RR{
		&dns.A{A: net.IP{1, 2, 3, 4}},
	}}

	testCases := map[string]struct {
		response *dns.Msg
	}{
		"nil response and nil error": {},
		"response and error": {
			response: dummyResponse,
		},
		"response and no error": {
			response: dummyResponse,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			writer := NewWriter()

			err := writer.WriteMsg(testCase.response)

			assert.Equal(t, testCase.response, writer.Response)
			assert.NoError(t, err)
		})
	}
}
