package server

import "github.com/miekg/dns"

type testWriter struct {
	dns.ResponseWriter
	writeErrToReturn error
	responseWritten  *dns.Msg
}

func (w *testWriter) WriteMsg(response *dns.Msg) error {
	w.responseWritten = response
	return w.writeErrToReturn
}
