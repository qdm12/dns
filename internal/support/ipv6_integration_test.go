//go:build integration
// +build integration

package support

import (
	"context"
	"net/netip"
	"testing"
)

func Test_IPv6(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cloudflareIPv6AddrPort := netip.MustParseAddrPort("[2606:4700:4700::1111]:443")

	ipv6Supported, err := IPv6(ctx, cloudflareIPv6AddrPort)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("IPv6 supported:", ipv6Supported)
}
