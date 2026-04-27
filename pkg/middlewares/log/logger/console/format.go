package console

import (
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

func formatError(requestID uint16, errString string) string {
	return "id: " + strconv.FormatUint(uint64(requestID), 10) + "; error: " + errString
}

func formatRequestResponse(request, response *dns.Msg) string {
	requestString := formatRequest(request)
	responseString := formatResponse(response)
	return requestString + " => " + responseString
}

func formatRequest(request *dns.Msg) (s string) {
	s = "id: " + strconv.FormatUint(uint64(request.Id), 10)

	questionStrings := make([]string, len(request.Question))
	for i, question := range request.Question {
		questionStrings[i] = question.Name + " " +
			dns.Class(question.Qclass).String() + " " +
			dns.Type(question.Qtype).String()
	}

	switch len(questionStrings) {
	case 0:
		s += "; no question"
	case 1:
		s += "; question: " + questionStrings[0]
	default:
		s += "; questions: " + strings.Join(questionStrings, ", ")
	}

	return s
}

func formatResponse(response *dns.Msg) (s string) {
	s = "id: " + strconv.FormatUint(uint64(response.Id), 10)

	answerStrings := make([]string, len(response.Answer))
	for i, answer := range response.Answer {
		answerStrings[i] = rrToString(answer)
	}

	switch len(answerStrings) {
	case 0:
		s += "; no answer"
	case 1:
		s += "; answer: " + rrToString(response.Answer[0])
	default:
		s += "; answers: [\n  " +
			strings.Join(answerStrings, ",\n  ") + "\n]"
	}

	return s
}

func rrToString(answer dns.RR) (responseStr string) {
	responseStr = answer.String()
	responseStr = strings.TrimPrefix(responseStr, "\t")
	responseStr = strings.ReplaceAll(responseStr, "\t", " ")
	return responseStr
}
