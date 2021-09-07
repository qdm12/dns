package noop

import "github.com/miekg/dns"

type Formatter struct{}

func New() *Formatter {
	return &Formatter{}
}

func (f *Formatter) Request(request *dns.Msg) string {
	return ""
}
func (f *Formatter) Response(response *dns.Msg) string {
	return ""
}
func (f *Formatter) RequestResponse(request, response *dns.Msg) string {
	return ""
}
func (f *Formatter) Error(requestID uint16, errString string) string {
	return ""
}
