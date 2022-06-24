package doh

import (
	"github.com/qdm12/dns/v2/pkg/provider"
)

type Picker interface {
	DoHServer(servers []provider.DoHServer) provider.DoHServer
}
