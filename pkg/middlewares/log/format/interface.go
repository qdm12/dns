package format

import (
	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/middlewares/log/format/console"
	"github.com/qdm12/dns/pkg/middlewares/log/format/noop"
)

var (
	_ Interface = (*console.Formatter)(nil)
	_ Interface = (*noop.Formatter)(nil)
)

type Interface interface {
	Request(request *dns.Msg) string
	Response(response *dns.Msg) string
	RequestResponse(request, response *dns.Msg) string
	Error(requestID uint16, errString string) string
}
