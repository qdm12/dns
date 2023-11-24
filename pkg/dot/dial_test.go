package dot

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_pickNameAddress(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	picker := NewMockPicker(ctrl)
	servers := []provider.DoTServer{
		provider.Cloudflare().DoT,
		provider.Google().DoT,
	}
	const ipv6 = true

	picker.EXPECT().DoTServer(servers).Return(servers[0])
	picker.EXPECT().DoTAddrPort(servers[0], ipv6).Return(servers[0].IPv6[0])

	name, address := pickNameAddress(picker, servers, ipv6)

	assert.Equal(t, "cloudflare-dns.com", name)
	assert.Equal(t, "[2606:4700:4700::1111]:853", address)
}
