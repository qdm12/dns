package dnssec

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type refusedHandler struct {
	calls int
}

func (h *refusedHandler) ServeDNS(w dns.ResponseWriter, request *dns.Msg) {
	h.calls++
	response := new(dns.Msg)
	response.SetRcode(request, dns.RcodeRefused)
	_ = w.WriteMsg(response)
}

func Test_Validate_RefusedPassThrough(t *testing.T) {
	t.Parallel()

	handler := new(refusedHandler)
	request := new(dns.Msg)
	request.SetQuestion("graph.instagram.com.", dns.TypeA)

	response, err := Validate(request, handler)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, dns.RcodeRefused, response.Rcode)
	assert.Equal(t, 1, handler.calls)
}
