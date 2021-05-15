package stateful

import (
	"github.com/miekg/dns"
)

// Writer wraps the dns writer in order to report
// the dns response written and eventual error.
type Writer struct {
	dns.ResponseWriter
	Response *dns.Msg
	WriteErr error
}

func (w *Writer) WriteMsg(response *dns.Msg) error {
	w.Response = response
	w.WriteErr = w.ResponseWriter.WriteMsg(response)
	return w.WriteErr
}

func NewWriter(w dns.ResponseWriter) *Writer {
	return &Writer{
		ResponseWriter: w,
	}
}
