package nameserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetDNSServers(t *testing.T) {
	servers := GetDNSServers()
	assert.NotEmpty(t, servers)
	for _, server := range servers {
		assert.True(t, server.IsValid())
	}
}
