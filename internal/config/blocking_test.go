package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"inet.af/netaddr"
)

func Test_getPrivateIPPrefixes(t *testing.T) {
	t.Parallel()

	expectedPrivateIPNets := []netaddr.IPPrefix{
		{IP: netaddr.IPv4(127, 0, 0, 1), Bits: 8},
		{IP: netaddr.IPv4(10, 0, 0, 0), Bits: 8},
		{IP: netaddr.IPv4(172, 16, 0, 0), Bits: 12},
		{IP: netaddr.IPv4(192, 168, 0, 0), Bits: 16},
		{IP: netaddr.IPv4(169, 254, 0, 0), Bits: 16},
		{ // ::1/128
			IP: netaddr.IPv6Raw([16]byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}),
			Bits: 128},
		{ // ::fc00::/7
			IP:   netaddr.IPv6Raw([16]byte{0xfc, 0x00, 0, 0, 0, 0}),
			Bits: 7},
		{ // fe80::/10
			IP:   netaddr.IPv6Raw([16]byte{0xfe, 0x80, 0, 0, 0, 0}),
			Bits: 10},
		{ // ::ffff:7F00:1/104
			IP: netaddr.IPv6Raw([16]byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0x7f, 0x00, 0, 1}),
			Bits: 104},
		{ // ::ffff:a00:0/104
			IP: netaddr.IPv6Raw([16]byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0x0a, 0x00, 0, 0}),
			Bits: 104},
		{ // ::ffff:ac10:0/108
			IP: netaddr.IPv6Raw([16]byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0xac, 0x10, 0, 0}),
			Bits: 108},
		{ // ::ffff:c0a8:0/112
			IP: netaddr.IPv6Raw([16]byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0xc0, 0xa8, 0, 0}),
			Bits: 112},
		{ // ::ffff:a9fe:0/112
			IP: netaddr.IPv6Raw([16]byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0xa9, 0xfe, 0, 0}),
			Bits: 112},
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
		"::ffff:7f00:1/104", // 127.0.0.1/8
		"::ffff:a00:0/104",  // 10.0.0.0/8
		"::ffff:ac10:0/108", // 172.16.0.0/12
		"::ffff:c0a8:0/112", // 192.168.0.0/16
		"::ffff:a9fe:0/112", // 169.254.0.0/16
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
