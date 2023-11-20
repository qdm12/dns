package mockhelp

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/miekg/dns"
	"golang.org/x/exp/maps"
)

func NewMatcherResponse(settings MatcherResponseSettings) *MatcherResponse {
	err := settings.validate()
	if err != nil {
		panic(err)
	}

	hasAnswerTypes := make(map[uint16]struct{})
	for _, answerType := range settings.OnlyHasAnswerTypes {
		hasAnswerTypes[answerType] = struct{}{}
	}

	ignoreAnswerTypes := make(map[uint16]struct{})
	for _, answerType := range settings.IgnoreAnswerTypes {
		ignoreAnswerTypes[answerType] = struct{}{}
	}

	return &MatcherResponse{
		response:           settings.Response.Copy(),
		onlyHasAnswerTypes: hasAnswerTypes,
		ignoreAnswerTypes:  ignoreAnswerTypes,
	}
}

type MatcherResponseSettings struct {
	// Response is the base expected response.
	// By default, the following are ignored:
	// - MsgHdr.Id field
	// - TTL field set in the answer headers
	// - Rdlength field set in the answer headers
	// - A and AAAA IP values, except checking both are empty or both non empty.
	// - Extra EDNS0 Padding length.
	Response *dns.Msg
	// OnlyHasAnswerTypes checks the response has at least one answer
	// for each of the types specified and no other answer types.
	OnlyHasAnswerTypes []uint16
	// IgnoreAnswerTypes removes the answers matching one of the types
	// specified from the received AND expected answers.
	IgnoreAnswerTypes []uint16
}

var (
	errResponseNotSet     = errors.New("response is not set")
	errAnswerTypeNotValid = errors.New("answer type is not valid")
)

func (m MatcherResponseSettings) validate() (err error) {
	if m.Response == nil {
		return fmt.Errorf("%w", errResponseNotSet)
	}

	for _, answerType := range m.OnlyHasAnswerTypes {
		_, ok := dns.TypeToString[answerType]
		if !ok {
			return fmt.Errorf("only has answer types: %w: %d",
				errAnswerTypeNotValid, answerType)
		}
	}

	for _, answerType := range m.IgnoreAnswerTypes {
		_, ok := dns.TypeToString[answerType]
		if !ok {
			return fmt.Errorf("ignore answer types: %w: %d",
				errAnswerTypeNotValid, answerType)
		}
	}

	return nil
}

// MatcherResponse matches a message with a response
// and ignores the following:
// - MsgHdr.Id field
// - TTL field set in the answer headers
// - Rdlength field set in the answer headers
// - A and AAAA IP values, except checking both are empty or both non empty.
// - Extra EDNS0 Padding length.
// See MatcherResponseSettings for additional settings.
type MatcherResponse struct {
	mismatchReason     string
	response           *dns.Msg
	onlyHasAnswerTypes map[uint16]struct{}
	ignoreAnswerTypes  map[uint16]struct{}
}

func (m *MatcherResponse) String() string {
	return m.mismatchReason
}

func (m *MatcherResponse) Matches(x interface{}) bool {
	msg, ok := x.(*dns.Msg)
	if !ok {
		m.mismatchReason = "not a *dns.Msg"
		return false
	}

	// Copy since these are mutated by this function
	received := msg.Copy()
	expected := m.response.Copy()

	m.mismatchReason = checkOnlyHasAnswers(received, m.onlyHasAnswerTypes)
	if m.mismatchReason != "" {
		return false
	}

	filterAnswers(expected, m.ignoreAnswerTypes)
	filterAnswers(received, m.ignoreAnswerTypes)

	m.mismatchReason = checkAnswers(expected, received)
	if m.mismatchReason != "" {
		return false
	}

	m.mismatchReason = checkExtras(expected, received)
	if m.mismatchReason != "" {
		return false
	}

	expectedPacked, _ := expected.Pack()
	receivedPacked, _ := received.Pack()
	match := bytes.Equal(expectedPacked, receivedPacked)
	if !match {
		m.mismatchReason = fmt.Sprintf("packed mismatch:\n"+
			"==> expected: %x\n==> received: %x\n",
			expectedPacked, receivedPacked)
	}

	return match
}

func checkOnlyHasAnswers(response *dns.Msg,
	onlyHasAnswerTypes map[uint16]struct{}) (
	mismatchReason string) {
	answerTypesNotPresent := maps.Clone(onlyHasAnswerTypes)
	for _, answer := range response.Answer {
		answerType := answer.Header().Rrtype
		_, expectedType := onlyHasAnswerTypes[answerType]
		if !expectedType {
			return fmt.Sprintf("unexpected answer type: %s",
				dns.TypeToString[answerType])
		}
		delete(answerTypesNotPresent, answerType)
	}

	if len(answerTypesNotPresent) == 0 {
		return ""
	}

	answerTypesMissing := make([]string, 0, len(answerTypesNotPresent))
	for answerType := range answerTypesNotPresent {
		answerTypesMissing = append(answerTypesMissing,
			dns.TypeToString[answerType])
	}
	return fmt.Sprintf("missing answer types: %s",
		strings.Join(answerTypesMissing, ", "))
}

func filterAnswers(response *dns.Msg,
	ignoreAnswerTypes map[uint16]struct{}) {
	filteredAnswers := make([]dns.RR, 0, len(response.Answer))
	for _, answer := range response.Answer {
		answerType := answer.Header().Rrtype
		_, ignored := ignoreAnswerTypes[answerType]
		if ignored {
			continue
		}
		filteredAnswers = append(filteredAnswers, answer)
	}
	response.Answer = filteredAnswers
}

func checkAnswers(expected, received *dns.Msg) (
	mismatchReason string) {
	if len(received.Answer) != len(expected.Answer) {
		return fmt.Sprintf("answers count mismatch: "+
			"expected %d, received %d",
			len(expected.Answer), len(received.Answer))
	}

	// Clear randomly set fields
	received.MsgHdr.Id = 0
	expected.MsgHdr.Id = 0

	for i := range received.Answer {
		if !answersAreEqual(expected.Answer[i], received.Answer[i]) {
			return fmt.Sprintf("answer %d of %d mismatch:\n"+
				"==> expected: %s\nreceived: %s\n",
				i+1, len(expected.Answer), expected.Answer[i], received.Answer[i])
		}
	}

	return ""
}

func checkExtras(expected, received *dns.Msg) (
	mismatchReason string) {
	if len(received.Extra) != len(expected.Extra) {
		return fmt.Sprintf("extra count mismatch: "+
			"expected %d, received %d",
			len(expected.Extra), len(received.Extra))
	}

	for i := range received.Extra {
		if !extrasAreEqual(expected.Extra[i], received.Extra[i]) {
			return fmt.Sprintf("extra %d of %d mismatch:\n"+
				"==> expected: %s\n==> received: %s",
				i+1, len(expected.Extra), expected.Extra[i], received.Extra[i])
		}
	}

	return ""
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
	default:
		panic(fmt.Sprintf("unexpected answer type: %s",
			dns.TypeToString[receivedHeader.Rrtype]))
	}
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
	// Clear values for packed bytes comparison in caller function
	receivedAnswer.A = nil
	expectedAnswer.A = nil
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
	// Clear values for packed bytes comparison in caller function
	receivedAnswer.AAAA = nil
	expectedAnswer.AAAA = nil
	return true
}

func extrasAreEqual(expected, actual dns.RR) (equal bool) {
	receivedHeader := actual.Header()
	expectedHeader := expected.Header()

	if receivedHeader.Rrtype != expectedHeader.Rrtype {
		return false
	}

	switch receivedHeader.Rrtype {
	case dns.TypeOPT:
		return extrasOPTEqual(expected, actual)
	default:
		panic(fmt.Sprintf("unexpected extra type: %s",
			dns.TypeToString[receivedHeader.Rrtype]))
	}
}

func extrasOPTEqual(expected, actual dns.RR) (equal bool) {
	receivedExtra, ok := actual.(*dns.OPT)
	if !ok {
		return false
	}
	expectedExtra, ok := expected.(*dns.OPT)
	if !ok {
		return false
	}

	if len(expectedExtra.Option) != len(receivedExtra.Option) {
		return false
	}

	for i, expectedOption := range expectedExtra.Option {
		receivedOption := receivedExtra.Option[i]
		switch expectedOption.Option() {
		case dns.EDNS0PADDING:
			return edns0PaddingEqual(expectedOption, receivedOption)
		default:
			panic(fmt.Sprintf("unexpected option type: %T", expectedOption))
		}
	}

	return true
}

func edns0PaddingEqual(expected, actual dns.EDNS0) (equal bool) {
	receivedPadding, ok := actual.(*dns.EDNS0_PADDING)
	if !ok {
		return false
	}
	expectedPadding, ok := expected.(*dns.EDNS0_PADDING)
	if !ok {
		return false
	}

	// Ignore the EDNS0 Padding which can change depending on the
	// rest of the message due to compression + encryption.
	expectedPadding.Padding = []byte{0}
	receivedPadding.Padding = []byte{0}
	return true
}
