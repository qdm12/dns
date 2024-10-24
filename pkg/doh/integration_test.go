//go:build integration
// +build integration

package doh

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Resolver(t *testing.T) {
	t.Parallel()

	settings := Settings{}
	dialer, err := New(settings)
	require.NoError(t, err)

	const hostname = "google.com"

	resolver := &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial:         dialer.Dial,
	}

	ips, err := resolver.LookupIPAddr(context.Background(), hostname)

	require.NoError(t, err)
	require.NotEmpty(t, ips)
	t.Logf("resolved %s to: %v", hostname, ips)
}
