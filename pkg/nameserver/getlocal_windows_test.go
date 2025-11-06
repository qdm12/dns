package nameserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetDNSServers(t *testing.T) {
	servers, err := GetDNSServers()
	require.NoError(t, err)
	assert.NotEmpty(t, servers)
	for _, server := range servers {
		assert.True(t, server.IsValid())
	}
}
