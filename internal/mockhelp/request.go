package mockhelp

import (
	"bytes"

	"github.com/miekg/dns"
)

func NewMatcherRequest(request *dns.Msg) *MatcherRequest {
	return &MatcherRequest{
		request: request.Copy(),
	}
}

type MatcherRequest struct {
	request *dns.Msg
}

func (m *MatcherRequest) String() string {
	return m.request.String() + " [ignoring .MsgHdr.Id]"
}

func (m *MatcherRequest) Matches(x interface{}) bool {
	msg, ok := x.(*dns.Msg)
	if !ok {
		return false
	}

	expected := m.request.Copy()
	expected.MsgHdr.Id = 0
	expectedPacked, _ := expected.Pack()

	received := msg.Copy()
	received.MsgHdr.Id = 0
	receivedPacked, _ := received.Pack()

	return bytes.Equal(expectedPacked, receivedPacked)
}
