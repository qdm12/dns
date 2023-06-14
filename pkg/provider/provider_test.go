package provider

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Provider_JSON(t *testing.T) {
	t.Parallel()

	provider := Google()

	providerJSON, err := json.Marshal(provider)
	require.NoError(t, err)

	const expectedProviderJSON = `{"name":"Google",` +
		`"dns":{"ipv4":["8.8.8.8","8.8.4.4"],"ipv6":["2001:4860:4860::8888","2001:4860:4860::8844"]},` +
		`"dot":{"ipv4":["8.8.8.8","8.8.4.4"],"ipv6":["2001:4860:4860::8888","2001:4860:4860::8844"],` +
		`"name":"dns.google","port":853},` +
		`"doh":{"url":"https://dns.google/dns-query"}}`
	assert.Equal(t, expectedProviderJSON, string(providerJSON))

	var decodedProvider Provider
	err = json.Unmarshal(providerJSON, &decodedProvider)
	require.NoError(t, err)
	assert.Equal(t, provider, decodedProvider)
}