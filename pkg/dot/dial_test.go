package dot

import (
	"testing"

	"github.com/qdm12/dns/v2/internal/picker"
	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_pickNameAddress(t *testing.T) {
	t.Parallel()

	picker := picker.New()
	servers := []provider.DoTServer{
		provider.Cloudflare().DoT,
		provider.Google().DoT,
	}
	const ipv6 = true

	name, address := pickNameAddress(picker, servers, ipv6)

	found := false
	for _, server := range servers {
		if server.Name != name {
			continue
		}
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
