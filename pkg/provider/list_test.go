package provider

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_All(t *testing.T) {
	t.Parallel()

	t.Logf("%#v", netip.MustParseAddr("2620:119:35::35").As16())

	providers := All()
	assert.Len(t, providers, 16)

	for _, provider := range providers {
		errMessage := "for provider " + provider.DoT.Name

		assert.NotEmpty(t, provider.DNS.IPv4, errMessage)
		assert.NotEmpty(t, provider.DNS.IPv6, errMessage)

		assert.NotEmpty(t, provider.DoT.IPv4, errMessage)
		assert.NotNil(t, provider.DoT.IPv6, errMessage)
		assert.NotEmpty(t, provider.DoT.Name, errMessage)
		assert.NotZero(t, provider.DoT.Port, errMessage)

		assert.NotNil(t, provider.DoH, errMessage)
	}
}
