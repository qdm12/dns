package noop

import (
	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/middlewares/filter/update"
)

type Filter struct{}

func New() *Filter {
	return new(Filter)
}

func (f *Filter) FilterRequest(_ *dns.Msg) (_ bool)  { return false }
func (f *Filter) FilterResponse(_ *dns.Msg) (_ bool) { return false }
func (f *Filter) Update(_ update.Settings)           {}
