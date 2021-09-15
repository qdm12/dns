package console

import "github.com/miekg/dns"

func (f *Formatter) RequestResponse(request, response *dns.Msg) string {
	requestString := f.Request(request)
	responseString := f.Response(response)
	return requestString + " => " + responseString
}
