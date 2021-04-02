package doh

import (
	"github.com/qdm12/dns/pkg/provider"
	"github.com/qdm12/golibs/crypto/random/hashmap"
)

type picker struct {
	rand hashmap.Rand
}

func newPicker() *picker {
	return &picker{
		rand: hashmap.New(),
	}
}

func (p *picker) DoHServer(servers []provider.DoHServer) provider.DoHServer {
	index := 0
	if nServers := len(servers); nServers > 1 {
		index = p.rand.Intn(nServers)
	}
	return servers[index]
}
