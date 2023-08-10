package stateful

import (
	"net"

	"github.com/miekg/dns"
)

// Writer is a stateful writer with the Response field
// set when WriteMsg is called. Only the WriteMsg method
// is implemented, calls to the other methods will panic.
type Writer struct {
	Response *dns.Msg
}

// NewWriter creates a new stateful writer.
func NewWriter() *Writer {
	return &Writer{}
}

// WriteMsg sets the Response field of the Writer
// to the given response message and always returns
// a nil error.
func (w *Writer) WriteMsg(response *dns.Msg) error {
	w.Response = response
	return nil
}

// LocalAddr will panic if called.
func (w *Writer) LocalAddr() net.Addr {
	panic("not implemented")
}

// RemoteAddr will panic if called.
func (w *Writer) RemoteAddr() net.Addr {
	panic("not implemented")
}

// Write will panic if called.
func (w *Writer) Write([]byte) (int, error) {
	panic("not implemented")
}

// Close will panic if called.
func (w *Writer) Close() error {
	panic("not implemented")
}

// TsigStatus will panic if called.
func (w *Writer) TsigStatus() error {
	panic("not implemented")
}

// TsigTimersOnly will panic if called.
func (w *Writer) TsigTimersOnly(bool) {
	panic("not implemented")
}

// Hijack will panic if called.
func (w *Writer) Hijack() {
	panic("not implemented")
}
