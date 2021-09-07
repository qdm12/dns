package console

import "github.com/miekg/dns"

func (f *Formatter) RequestResponse(request, response *dns.Msg) string {
	requestString, ok := f.idToRequestString[request.Id]
	if !ok {
		requestString = f.Request(request)
	} else {
		delete(f.idToRequestString, request.Id)
	}

	responseString, ok := f.idToResponseString[response.Id]
	if !ok {
		responseString = f.Response(response)
	} else {
		delete(f.idToResponseString, response.Id)
	}

	return requestString + " => " + responseString
}
