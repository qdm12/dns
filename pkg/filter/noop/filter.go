package noop

import "github.com/miekg/dns"

type Filter struct{}

func New() *Filter {
	return new(Filter)
}

func (f *Filter) FilterRequest(_ *dns.Msg) (_ bool)  { return false }
func (f *Filter) FilterResponse(_ *dns.Msg) (_ bool) { return false }
