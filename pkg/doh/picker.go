package doh

import (
	"math/rand"

	"github.com/qdm12/dns/pkg/provider"
	"github.com/qdm12/golibs/crypto/random/sources/maphash"
)

type picker struct {
	rand *rand.Rand
}

func newPicker() *picker {
	source := maphash.New()
	return &picker{
		rand: rand.New(source), //nolint:gosec
	}
}

func (p *picker) DoHServer(servers []provider.DoHServer) provider.DoHServer {
	index := 0
	if nServers := len(servers); nServers > 1 {
		index = p.rand.Intn(nServers)
	}
	return servers[index]
}
