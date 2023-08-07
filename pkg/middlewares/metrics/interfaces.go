package metrics

type Metrics interface {
	RequestsInc()
	QuestionsInc(class, qType string)
	RcodeInc(rcode string)
	AnswersInc(class, qType string)
	ResponsesInc()
	InFlightRequestsInc()
	InFlightRequestsDec()
}
