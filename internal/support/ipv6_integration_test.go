//go:build integration
// +build integration

package support

import (
	"context"
	"testing"
)

func Test_IPv6(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ipv6Supported, err := IPv6(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("IPv6 supported:", ipv6Supported)
}
