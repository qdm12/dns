// Package noop defines a No-Op metric implementation for DoT.
package noop

import (
	middleware "github.com/qdm12/dns/pkg/middlewares/metrics"
	middlewarenoop "github.com/qdm12/dns/pkg/middlewares/metrics/noop"
)

type middlewareInterface = middleware.Interface

type Metrics struct {
	middlewareInterface
}

func New() *Metrics {
	return &Metrics{
		middlewareInterface: middlewarenoop.New(),
	}
}

func (m *Metrics) DoTDialProviderInc(provider, outcome string) {}
func (m *Metrics) DoTDialAddressInc(address, outcome string)   {}
func (m *Metrics) DNSDialProviderInc(provider, outcome string) {}
func (m *Metrics) DNSDialAddressInc(address, outcome string)   {}
