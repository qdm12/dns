package server

import "github.com/miekg/dns"

func ptrTo[T any](value T) *T {
	return &value
}

type testWriter struct {
	dns.ResponseWriter
	writeErrToReturn error
	responseWritten  *dns.Msg
}

func (w *testWriter) WriteMsg(response *dns.Msg) error {
	w.responseWritten = response
	return w.writeErrToReturn
}
