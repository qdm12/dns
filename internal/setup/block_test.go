package setup

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getPrivateIPPrefixes(t *testing.T) {
	t.Parallel()

	expectedPrivateIPNets := []netip.Prefix{
		netip.PrefixFrom(netip.AddrFrom4([4]byte{127, 0, 0, 1}), 8),
		netip.PrefixFrom(netip.AddrFrom4([4]byte{10, 0, 0, 0}), 8),
		netip.PrefixFrom(netip.AddrFrom4([4]byte{172, 16, 0, 0}), 12),
		netip.PrefixFrom(netip.AddrFrom4([4]byte{192, 168, 0, 0}), 16),
		netip.PrefixFrom(netip.AddrFrom4([4]byte{169, 254, 0, 0}), 16),
		// ::1/128
		netip.PrefixFrom(netip.AddrFrom16([16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}), 128),
		// ::fc00::/7
		netip.PrefixFrom(netip.AddrFrom16([16]byte{0xfc, 0x00, 0, 0, 0, 0}), 7),
		// fe80::/10
		netip.PrefixFrom(netip.AddrFrom16([16]byte{0xfe, 0x80, 0, 0, 0, 0}), 10),
		// ::ffff:7F00:1/104
		netip.PrefixFrom(netip.AddrFrom16([16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0x7f, 0x00, 0, 1}), 104),
		// ::ffff:a00:0/104
		netip.PrefixFrom(netip.AddrFrom16([16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0x0a, 0x00, 0, 0}), 104),
		// ::ffff:ac10:0/108
		netip.PrefixFrom(netip.AddrFrom16([16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0xac, 0x10, 0, 0}), 108),
		// ::ffff:c0a8:0/112
		netip.PrefixFrom(netip.AddrFrom16([16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0xc0, 0xa8, 0, 0}), 112),
		// ::ffff:a9fe:0/112
		netip.PrefixFrom(netip.AddrFrom16([16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0xa9, 0xfe, 0, 0}), 112),
	}

	expectedStrings := []string{
		// IPv4 private addresses
		"127.0.0.1/8",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
		// IPv6 private addresses
		"::1/128",
		"fc00::/7",
		"fe80::/10",
		// Private IPv4 addresses wrapped in IPv6
		"::ffff:127.0.0.1/104",   // 127.0.0.1/8
		"::ffff:10.0.0.0/104",    // 10.0.0.0/8
		"::ffff:172.16.0.0/108",  // 172.16.0.0/12
		"::ffff:192.168.0.0/112", // 192.168.0.0/16
		"::ffff:169.254.0.0/112", // 169.254.0.0/16
	}

	privateIPNets, err := getPrivateIPPrefixes()
	require.NoError(t, err)

	assert.Equal(t, len(expectedPrivateIPNets), len(privateIPNets))
	assert.Equal(t, len(expectedStrings), len(privateIPNets))

	for i := range privateIPNets {
		assert.Equal(t, expectedPrivateIPNets[i], privateIPNets[i])
		assert.Equal(t, expectedStrings[i], privateIPNets[i].String())
	}
}
