package format

import (
	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/middlewares/log/format/console"
	"github.com/qdm12/dns/pkg/middlewares/log/format/noop"
)

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE . Interface

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
