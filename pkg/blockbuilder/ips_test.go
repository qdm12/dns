package blockbuilder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Builder_IPs(t *testing.T) { //nolint:cyclop
	t.Parallel()
	type blockParams struct {
		blocked   bool
		content   []byte
		clientErr error
	}
	tests := map[string]struct {
		malicious                   blockParams
		ads                         blockParams
		surveillance                blockParams
		allowedIPs                  []netip.Addr
		additionalBlockedIPs        []netip.Addr
		allowedIPPrefixes           []netip.Prefix
		additionalBlockedIPPrefixes []netip.Prefix
		blockedIPs                  []string // string format for easier comparison
		blockedIPPrefixes           []string // string format for easier comparison
		errsString                  []string // string format for easier comparison
	}{
		"nothing blocked": {},
		"only malicious blocked": {
			malicious: blockParams{
				blocked: true,
				content: []byte("1.2.3.4\n99.99.99.99/24"),
			},
			blockedIPs:        []string{"1.2.3.4"},
			blockedIPPrefixes: []string{"99.99.99.99/24"},
		},
		"all blocked with some duplicates": {
			malicious: blockParams{
				blocked: true,
				content: []byte("1.2.3.4\n66.67.68.10/28"),
			},
			ads: blockParams{
				blocked: true,
				content: []byte("1.2.3.4\n254.254.254.1"),
			},
			surveillance: blockParams{
				blocked: true,
				content: []byte("254.254.254.1\n1.2.3.4"),
			},
			blockedIPs:        []string{"1.2.3.4", "254.254.254.1"},
			blockedIPPrefixes: []string{"66.67.68.10/28"},
		},
		"all blocked with one errored": {
			malicious: blockParams{
				blocked: true,
				content: []byte("1.2.3.4\n66.67.68.10/28"),
			},
			ads: blockParams{
				blocked: true,
				content: []byte("1.2.3.4\n254.254.254.1"),
			},
			surveillance: blockParams{
				blocked:   true,
				clientErr: fmt.Errorf("surveillance error"),
			},
			blockedIPs:        []string{"1.2.3.4", "254.254.254.1"},
			blockedIPPrefixes: []string{"66.67.68.10/28"},
			errsString: []string{
				`Get "https://raw.githubusercontent.com/qdm12/files/master/surveillance-ips.updated": surveillance error`,
			},
		},
		"blocked with private addresses": {
			malicious: blockParams{
				blocked: true,
				content: []byte("1.2.3.4\n66.67.68.10/28"),
			},
			ads: blockParams{
				blocked: true,
				content: []byte("254.254.254.1"),
			},
			additionalBlockedIPs: []netip.Addr{netip.AddrFrom4([4]byte{254, 254, 254, 1})},
			additionalBlockedIPPrefixes: []netip.Prefix{
				netip.PrefixFrom(netip.AddrFrom4([4]byte{55, 55, 55, 0}), 24),
			},
			blockedIPs:        []string{"1.2.3.4", "254.254.254.1"},
			blockedIPPrefixes: []string{"66.67.68.10/28", "55.55.55.0/24"},
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			clientCalls := struct {
				m map[string]int
				sync.Mutex
			}{
				m: make(map[string]int),
			}
			if tc.malicious.blocked {
				clientCalls.m[maliciousBlockListIPsURL] = 0
			}
			if tc.ads.blocked {
				clientCalls.m[adsBlockListIPsURL] = 0
			}
			if tc.surveillance.blocked {
				clientCalls.m[surveillanceBlockListIPsURL] = 0
			}

			errUnknownURL := errors.New("unknown URL")

			client := &http.Client{
				Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					url := r.URL.String()
					clientCalls.Lock()
					defer clientCalls.Unlock()
					if _, ok := clientCalls.m[url]; !ok {
						return nil, fmt.Errorf("%w: %q", errUnknownURL, url)
					}
					clientCalls.m[url]++
					var body []byte
					var err error
					switch url {
					case maliciousBlockListIPsURL:
						body = tc.malicious.content
						err = tc.malicious.clientErr
					case adsBlockListIPsURL:
						body = tc.ads.content
						err = tc.ads.clientErr
					case surveillanceBlockListIPsURL:
						body = tc.surveillance.content
						err = tc.surveillance.clientErr
					default: // just in case if the test is badly written
						return nil, fmt.Errorf("%w: %q", errUnknownURL, url)
					}
					if err != nil {
						return nil, err
					}
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(body)),
					}, nil
				}),
			}

			settings := Settings{Client: client}
			builder, err := New(settings)
			require.NoError(t, err)

			blockedIPs, blockedIPPrefixes, errs := builder.buildIPs(ctx,
				tc.malicious.blocked, tc.ads.blocked, tc.surveillance.blocked,
				tc.allowedIPs, tc.additionalBlockedIPs,
				tc.allowedIPPrefixes, tc.additionalBlockedIPPrefixes)

			assert.ElementsMatch(t, tc.blockedIPs, convertIPsToString(blockedIPs))
			assert.ElementsMatch(t, tc.blockedIPPrefixes, convertIPPrefixesToString(blockedIPPrefixes))
			assert.ElementsMatch(t, tc.errsString, convertErrorsToString(errs))
		})
	}
}
