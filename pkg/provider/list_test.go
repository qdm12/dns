package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_All(t *testing.T) {
	t.Parallel()
	providers := All()
	assert.Len(t, providers, 15)

	for _, provider := range providers {
		errMessage := "for provider " + provider.DoT().Name

		dnsServer := provider.DNS()
		assert.NotEmpty(t, dnsServer.IPv4, errMessage)
		assert.NotEmpty(t, dnsServer.IPv6, errMessage)

		dotServer := provider.DoT()
		assert.NotEmpty(t, dotServer.IPv4, errMessage)
		assert.NotNil(t, dotServer.IPv6, errMessage)
		assert.NotEmpty(t, dotServer.Name, errMessage)
		assert.NotZero(t, dotServer.Port, errMessage)

		dohServer := provider.DoH()
		assert.NotNil(t, dohServer.URL, errMessage)
	}
}
