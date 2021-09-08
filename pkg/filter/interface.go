package filter

import (
	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/filter/mapfilter"
	"github.com/qdm12/dns/pkg/filter/noop"
	"github.com/qdm12/dns/pkg/filter/update"
)

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE . Interface

var (
	_ Interface = (*mapfilter.Filter)(nil)
	_ Interface = (*noop.Filter)(nil)
)

type Interface interface {
	FilterRequest(request *dns.Msg) (blocked bool)
	FilterResponse(response *dns.Msg) (blocked bool)
	Update(settings update.Settings)
}
