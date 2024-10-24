package plain

import (
	"testing"

	"github.com/qdm12/dns/v2/internal/picker"
	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_pickAddress(t *testing.T) {
	t.Parallel()

	picker := picker.New()
	servers := []provider.PlainServer{
		provider.Cloudflare().Plain,
		provider.Google().Plain,
	}
	const ipv6 = true

	address := pickAddress(picker, servers, ipv6)

	found := false
	for _, server := range servers {
		ips := server.IPv4
		if ipv6 {
			ips = append(ips, server.IPv6...)
		}
		for _, addrPort := range ips {
			if addrPort.String() == address {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	assert.True(t, found)
}
