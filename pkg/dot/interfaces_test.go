package dot

type middlewareMetrics interface { //nolint:unused
	RequestsInc()
	QuestionsInc(class, qType string)
	RcodeInc(rcode string)
	AnswersInc(class, qType string)
	ResponsesInc()
	InFlightRequestsInc()
	InFlightRequestsDec()
}
