package mockhelp

import (
	"bytes"

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
		if !answersAreEqual(expected.Answer[i], answer) {
			return false
		}
	}

	expectedPacked, _ := expected.Pack()
	receivedPacked, _ := received.Pack()

	return bytes.Equal(expectedPacked, receivedPacked)
}

func answersAreEqual(expected, actual dns.RR) (equal bool) {
	receivedHeader := actual.Header()
	expectedHeader := expected.Header()

	if receivedHeader.Rrtype != expectedHeader.Rrtype {
		return false
	}

	receivedHeader.Ttl, expectedHeader.Ttl = 0, 0
	receivedHeader.Rdlength, expectedHeader.Rdlength = 0, 0

	switch receivedHeader.Rrtype {
	case dns.TypeA:
		return answersAEqual(expected, actual)
	case dns.TypeAAAA:
		return answersAAAAEqual(expected, actual)
	}
	return false
}

func answersAEqual(expected, actual dns.RR) (equal bool) {
	receivedAnswer, ok := actual.(*dns.A)
	if !ok {
		return false
	}
	expectedAnswer, ok := expected.(*dns.A)
	if !ok {
		return false
	}
	// Only check the length
	receivedIP := receivedAnswer.A
	expectedIP := expectedAnswer.A
	if len(expectedIP) == 0 && len(receivedIP) > 0 {
		return false
	} else if len(expectedIP) > 0 && len(receivedIP) == 0 {
		return false
	}
	return true
}

func answersAAAAEqual(expected, actual dns.RR) (equal bool) {
	receivedAnswer, ok := actual.(*dns.AAAA)
	if !ok {
		return false
	}
	expectedAnswer, ok := expected.(*dns.AAAA)
	if !ok {
		return false
	}
	// Only check the length
	receivedIP := receivedAnswer.AAAA
	expectedIP := expectedAnswer.AAAA
	if len(expectedIP) == 0 && len(receivedIP) > 0 {
		return false
	} else if len(expectedIP) > 0 && len(receivedIP) == 0 {
		return false
	}
	return true
}
