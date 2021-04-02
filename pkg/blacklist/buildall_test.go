package blacklist

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_builder_All(t *testing.T) {
	t.Parallel()
	type blockParams struct {
		blocked            bool
		contentHostnames   []byte
		clientHostnamesErr error
		contentIps         []byte
		clientIpsErr       error
	}
	tests := map[string]struct {
		malicious                  blockParams
		ads                        blockParams
		surveillance               blockParams
		additionalBlockedHostnames []string
		allowedHostnames           []string
		additionalBlockedIPs       []net.IP
		additionalBlockedIPNets    []*net.IPNet
		blockedHostnames           []string
		blockedIPs                 []string // string format for easier comparison
		blockedIPNets              []string // string format for easier comparison
		errsString                 []string // string format for easier comparison
	}{
		"none blocked": {},
		"all blocked without lists": {
			malicious: blockParams{
				blocked: true,
			},
			ads: blockParams{
				blocked: true,
			},
			surveillance: blockParams{
				blocked: true,
			},
		},
		"all blocked with lists": {
			malicious: blockParams{
				blocked:          true,
				contentHostnames: []byte("malicious.com"),
				contentIps:       []byte("1.2.3.4"),
			},
			ads: blockParams{
				blocked:          true,
				contentHostnames: []byte("ads.com"),
				contentIps:       []byte("1.2.3.5"),
			},
			surveillance: blockParams{
				blocked:          true,
				contentHostnames: []byte("surveillance.com"),
				contentIps:       []byte("1.2.3.6"),
			},
			blockedHostnames: []string{"ads.com", "malicious.com", "surveillance.com"},
			blockedIPs:       []string{"1.2.3.4", "1.2.3.5", "1.2.3.6"},
		},
		"all blocked with allowed hostnames": {
			malicious: blockParams{
				blocked:          true,
				contentHostnames: []byte("malicious.com"),
				contentIps:       []byte("1.2.3.4"),
			},
			ads: blockParams{
				blocked:          true,
				contentHostnames: []byte("ads.com"),
				contentIps:       []byte("1.2.3.5"),
			},
			surveillance: blockParams{
				blocked:          true,
				contentHostnames: []byte("surveillance.com"),
				contentIps:       []byte("1.2.3.6"),
			},
			allowedHostnames: []string{"ads.com"},
			blockedHostnames: []string{"malicious.com", "surveillance.com"},
			blockedIPs:       []string{"1.2.3.4", "1.2.3.5", "1.2.3.6"},
		},
		"all blocked with additional blocked IP addresses": {
			malicious: blockParams{
				blocked:          true,
				contentHostnames: []byte("malicious.com"),
				contentIps:       []byte("1.2.3.4"),
			},
			ads: blockParams{
				blocked:          true,
				contentHostnames: []byte("ads.com"),
				contentIps:       []byte("1.2.3.5"),
			},
			surveillance: blockParams{
				blocked:          true,
				contentHostnames: []byte("surveillance.com"),
				contentIps:       []byte("1.2.3.6"),
			},
			additionalBlockedIPs: []net.IP{{1, 2, 3, 7}},
			blockedHostnames:     []string{"ads.com", "malicious.com", "surveillance.com"},
			blockedIPs:           []string{"1.2.3.4", "1.2.3.5", "1.2.3.6", "1.2.3.7"},
		},
		"all blocked with lists and one error": {
			malicious: blockParams{
				blocked:          true,
				contentHostnames: []byte("malicious.com"),
				contentIps:       []byte("1.2.3.4"),
			},
			ads: blockParams{
				blocked:            true,
				contentHostnames:   []byte("ads.com"),
				clientHostnamesErr: fmt.Errorf("ads error"),
				contentIps:         []byte("1.2.3.5"),
			},
			surveillance: blockParams{
				blocked:          true,
				contentHostnames: []byte("surveillance.com"),
				contentIps:       []byte("1.2.3.6"),
			},
			additionalBlockedIPs: []net.IP{{1, 2, 3, 7}},
			blockedHostnames:     []string{"malicious.com", "surveillance.com"},
			blockedIPs:           []string{"1.2.3.4", "1.2.3.5", "1.2.3.6", "1.2.3.7"},
			errsString: []string{
				`Get "https://raw.githubusercontent.com/qdm12/files/master/ads-hostnames.updated": ads error`,
			},
		},
		"all blocked with errors": {
			malicious: blockParams{
				blocked:            true,
				clientIpsErr:       fmt.Errorf("malicious ips"),
				clientHostnamesErr: fmt.Errorf("malicious hostnames"),
			},
			ads: blockParams{
				blocked:            true,
				clientIpsErr:       fmt.Errorf("ads ips"),
				clientHostnamesErr: fmt.Errorf("ads hostnames"),
			},
			surveillance: blockParams{
				blocked:            true,
				clientIpsErr:       fmt.Errorf("surveillance ips"),
				clientHostnamesErr: fmt.Errorf("surveillance hostnames"),
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
			if tc.malicious.blocked {
				clientCalls.m[maliciousBlockListIPsURL] = 0
				clientCalls.m[maliciousBlockListHostnamesURL] = 0
			}
			if tc.ads.blocked {
				clientCalls.m[adsBlockListIPsURL] = 0
				clientCalls.m[adsBlockListHostnamesURL] = 0
			}
			if tc.surveillance.blocked {
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
						body = tc.malicious.contentIps
						err = tc.malicious.clientIpsErr
					case maliciousBlockListHostnamesURL:
						body = tc.malicious.contentHostnames
						err = tc.malicious.clientHostnamesErr
					case adsBlockListIPsURL:
						body = tc.ads.contentIps
						err = tc.ads.clientIpsErr
					case adsBlockListHostnamesURL:
						body = tc.ads.contentHostnames
						err = tc.ads.clientHostnamesErr
					case surveillanceBlockListIPsURL:
						body = tc.surveillance.contentIps
						err = tc.surveillance.clientIpsErr
					case surveillanceBlockListHostnamesURL:
						body = tc.surveillance.contentHostnames
						err = tc.surveillance.clientHostnamesErr
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

			builder := NewBuilder(client)

			blockedHostnames, blockedIPs, blockedIPNets, errs := builder.All(ctx,
				tc.malicious.blocked, tc.ads.blocked, tc.surveillance.blocked,
				tc.additionalBlockedHostnames, tc.allowedHostnames,
				tc.additionalBlockedIPs, tc.additionalBlockedIPNets)

			assert.ElementsMatch(t, tc.blockedHostnames, blockedHostnames)
			assert.ElementsMatch(t, tc.blockedIPs, convertIPsToString(blockedIPs))
			assert.ElementsMatch(t, tc.blockedIPNets, convertIPNetsToString(blockedIPNets))
			assert.ElementsMatch(t, tc.errsString, convertErrorsToString(errs))

			for url, count := range clientCalls.m {
				assert.Equalf(t, 1, count, "for url %q", url)
			}
		})
	}
}
