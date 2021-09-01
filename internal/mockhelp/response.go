package mockhelp

import (
	"bytes"
	"net"

	"github.com/miekg/dns"
)

func NewMatcherResponse(response *dns.Msg) *MatcherResponse {
	return &MatcherResponse{
		response: response.Copy(),
	}
}

// MatcherResponse matches a message with a response
// and ignores the following fields:
// - MsgHdr.Id
// - TTL field set in the answer headers
// - Rdlength field set in the answer headers
// - A and AAAA actual IP values.
type MatcherResponse struct {
	response *dns.Msg
}

func (m *MatcherResponse) String() string {
	return m.response.String() + " [ignoring .MsgHdr.Id, .Answer[].Header.Ttl, .Answer[].Header.Rdlength]"
}
func (m *MatcherResponse) Matches(x interface{}) bool {
	msg, ok := x.(*dns.Msg)
	if !ok {
		return false
	}

	received := msg.Copy()
	expected := m.response.Copy()

	if len(received.Answer) != len(expected.Answer) {
		return false
	}

	// Clear randomly set fields
	received.MsgHdr.Id = 0
	expected.MsgHdr.Id = 0
	for i, answer := range received.Answer {
		receivedHeader := answer.Header()
		expectedHeader := expected.Answer[i].Header()

		if receivedHeader.Rrtype != expectedHeader.Rrtype {
			return false
		}

		receivedHeader.Ttl, expectedHeader.Ttl = 0, 0
		receivedHeader.Rdlength, expectedHeader.Rdlength = 0, 0

		var expectedIP, receivedIP net.IP
		switch receivedHeader.Rrtype {
		case dns.TypeA:
			receivedAnswer := received.Answer[i].(*dns.A)
			expectedAnswer := expected.Answer[i].(*dns.A)
			receivedIP = receivedAnswer.A
			expectedIP = expectedAnswer.A
			expectedAnswer.A = nil
			receivedAnswer.A = nil
		case dns.TypeAAAA:
			receivedAnswer := received.Answer[i].(*dns.AAAA)
			expectedAnswer := expected.Answer[i].(*dns.AAAA)
			receivedIP = receivedAnswer.AAAA
			expectedIP = expectedAnswer.AAAA
			expectedAnswer.AAAA = nil
			receivedAnswer.AAAA = nil
		}

		if len(expectedIP) == 0 && len(receivedIP) > 0 {
			return false
		} else if len(expectedIP) > 0 && len(receivedIP) == 0 {
			return false
		}
	}

	expectedPacked, _ := expected.Pack()
	receivedPacked, _ := received.Pack()

	return bytes.Equal(expectedPacked, receivedPacked)
}
