package blockbuilder

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"inet.af/netaddr"
)

func Test_builder_All(t *testing.T) {
	t.Parallel()
	type httpCase struct {
		content []byte
		err     error
	}
	tests := map[string]struct {
		settings          Settings
		maliciousHosts    httpCase
		maliciousIPs      httpCase
		adsHosts          httpCase
		adsIPs            httpCase
		surveillanceHosts httpCase
		surveillanceIPs   httpCase
		blockedHostnames  []string
		blockedIPs        []string // string format for easier comparison
		blockedIPPrefixes []string // string format for easier comparison
		errsString        []string // string format for easier comparison
	}{
		"none blocked": {},
		"all blocked without lists": {
			settings: Settings{
				BlockMalicious:    true,
				BlockAds:          true,
				BlockSurveillance: true,
			},
		},
		"all blocked with lists": {
			settings: Settings{
				BlockMalicious:    true,
				BlockAds:          true,
				BlockSurveillance: true,
			},
			maliciousHosts: httpCase{
				content: []byte("malicious.com"),
			},
			maliciousIPs: httpCase{
				content: []byte("1.2.3.4"),
			},
			adsHosts: httpCase{
				content: []byte("ads.com"),
			},
			adsIPs: httpCase{
				content: []byte("1.2.3.5"),
			},
			surveillanceHosts: httpCase{
				content: []byte("surveillance.com"),
			},
			surveillanceIPs: httpCase{
				content: []byte("1.2.3.6"),
			},
			blockedHostnames: []string{"ads.com", "malicious.com", "surveillance.com"},
			blockedIPs:       []string{"1.2.3.4", "1.2.3.5", "1.2.3.6"},
		},
		"all blocked with allowed hostnames": {
			settings: Settings{
				BlockMalicious:    true,
				BlockAds:          true,
				BlockSurveillance: true,
				AllowedHosts:      []string{"ads.com"},
			},
			maliciousHosts: httpCase{
				content: []byte("malicious.com"),
			},
			maliciousIPs: httpCase{
				content: []byte("1.2.3.4"),
			},
			adsHosts: httpCase{
				content: []byte("ads.com"),
			},
			adsIPs: httpCase{
				content: []byte("1.2.3.5"),
			},
			surveillanceHosts: httpCase{
				content: []byte("surveillance.com"),
			},
			surveillanceIPs: httpCase{
				content: []byte("1.2.3.6"),
			},
			blockedHostnames: []string{"malicious.com", "surveillance.com"},
			blockedIPs:       []string{"1.2.3.4", "1.2.3.5", "1.2.3.6"},
		},
		"blocked with additional blocked IP addresses": {
			settings: Settings{
				BlockMalicious: true,
				AddBlockedIPs:  []netaddr.IP{netaddr.IPv4(1, 2, 3, 7)},
			},
			maliciousHosts: httpCase{
				content: []byte("malicious.com"),
			},
			maliciousIPs: httpCase{
				content: []byte("1.2.3.4"),
			},
			blockedHostnames: []string{"malicious.com"},
			blockedIPs:       []string{"1.2.3.4", "1.2.3.7"},
		},
		"all blocked with lists and one error": {
			settings: Settings{
				BlockMalicious:    true,
				BlockAds:          true,
				BlockSurveillance: true,
			},
			maliciousHosts: httpCase{
				content: []byte("malicious.com"),
			},
			maliciousIPs: httpCase{
				content: []byte("1.2.3.4"),
			},
			adsHosts: httpCase{
				err: errors.New("ads error"),
			},
			adsIPs: httpCase{
				content: []byte("1.2.3.5"),
			},
			surveillanceHosts: httpCase{
				content: []byte("surveillance.com"),
			},
			surveillanceIPs: httpCase{
				content: []byte("1.2.3.6"),
			},
			blockedHostnames: []string{"malicious.com", "surveillance.com"},
			blockedIPs:       []string{"1.2.3.4", "1.2.3.5", "1.2.3.6"},
			errsString: []string{
				`Get "https://raw.githubusercontent.com/qdm12/files/master/ads-hostnames.updated": ads error`,
			},
		},
		"all blocked with errors": {
			settings: Settings{
				BlockMalicious:    true,
				BlockAds:          true,
				BlockSurveillance: true,
			},
			maliciousHosts: httpCase{
				err: errors.New("malicious hostnames"),
			},
			maliciousIPs: httpCase{
				err: errors.New("malicious ips"),
			},
			adsHosts: httpCase{
				err: errors.New("ads hostnames"),
			},
			adsIPs: httpCase{
				err: errors.New("ads ips"),
			},
			surveillanceHosts: httpCase{
				err: errors.New("surveillance hostnames"),
			},
			surveillanceIPs: httpCase{
				err: errors.New("surveillance ips"),
			},
			errsString: []string{
				`Get "https://raw.githubusercontent.com/qdm12/files/master/malicious-ips.updated": malicious ips`,
				`Get "https://raw.githubusercontent.com/qdm12/files/master/malicious-hostnames.updated": malicious hostnames`,
				`Get "https://raw.githubusercontent.com/qdm12/files/master/ads-ips.updated": ads ips`,
				`Get "https://raw.githubusercontent.com/qdm12/files/master/ads-hostnames.updated": ads hostnames`,
				`Get "https://raw.githubusercontent.com/qdm12/files/master/surveillance-ips.updated": surveillance ips`,
				`Get "https://raw.githubusercontent.com/qdm12/files/master/surveillance-hostnames.updated": surveillance hostnames`,
			},
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
			if tc.settings.BlockMalicious {
				clientCalls.m[maliciousBlockListIPsURL] = 0
				clientCalls.m[maliciousBlockListHostnamesURL] = 0
			}
			if tc.settings.BlockAds {
				clientCalls.m[adsBlockListIPsURL] = 0
				clientCalls.m[adsBlockListHostnamesURL] = 0
			}
			if tc.settings.BlockSurveillance {
				clientCalls.m[surveillanceBlockListIPsURL] = 0
				clientCalls.m[surveillanceBlockListHostnamesURL] = 0
			}

			client := &http.Client{
				Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					url := r.URL.String()
					clientCalls.Lock()
					defer clientCalls.Unlock()
					if _, ok := clientCalls.m[url]; !ok {
						t.Errorf("unknown URL %q", url)
						return nil, nil
					}
					clientCalls.m[url]++
					var body []byte
					var err error
					switch url {
					case maliciousBlockListIPsURL:
						body = tc.maliciousIPs.content
						err = tc.maliciousIPs.err
					case maliciousBlockListHostnamesURL:
						body = tc.maliciousHosts.content
						err = tc.maliciousHosts.err
					case adsBlockListIPsURL:
						body = tc.adsIPs.content
						err = tc.adsIPs.err
					case adsBlockListHostnamesURL:
						body = tc.adsHosts.content
						err = tc.adsHosts.err
					case surveillanceBlockListIPsURL:
						body = tc.surveillanceIPs.content
						err = tc.surveillanceIPs.err
					case surveillanceBlockListHostnamesURL:
						body = tc.surveillanceHosts.content
						err = tc.surveillanceHosts.err
					default: // just in case if the test is badly written
						t.Errorf("unknown URL %q", url)
						return nil, nil
					}
					if err != nil {
						return nil, err
					}
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewReader(body)),
					}, nil
				}),
			}

			builder := New(client)

			result := builder.All(ctx, tc.settings)

			assert.ElementsMatch(t, tc.blockedHostnames, result.BlockedHostnames)
			assert.ElementsMatch(t, tc.blockedIPs, convertIPsToString(result.BlockedIPs))
			assert.ElementsMatch(t, tc.blockedIPPrefixes, convertIPPrefixesToString(result.BlockedIPPrefixes))
			assert.ElementsMatch(t, tc.errsString, convertErrorsToString(result.Errors))

			for url, count := range clientCalls.m {
				assert.Equalf(t, 1, count, "for url %q", url)
			}
		})
	}
}
