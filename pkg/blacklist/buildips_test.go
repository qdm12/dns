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

	"github.com/qdm12/dns/internal/models"
	"github.com/stretchr/testify/assert"
)

func Test_builder_IPs(t *testing.T) {
	t.Parallel()
	type blockParams struct {
		blocked   bool
		content   []byte
		clientErr error
	}
	tests := map[string]struct {
		malicious               blockParams
		ads                     blockParams
		surveillance            blockParams
		additionalBlockedIPs    []net.IP
		additionalBlockedIPNets []*net.IPNet
		blockedIPs              []string // string format for easier comparison
		blockedIPNets           []string // string format for easier comparison
		errsString              []string // string format for easier comparison
	}{
		"nothing blocked": {},
		"only malicious blocked": {
			malicious: blockParams{
				blocked: true,
				content: []byte("1.2.3.4\n99.99.99.99/24"),
			},
			blockedIPs:    []string{"1.2.3.4"},
			blockedIPNets: []string{"99.99.99.0/24"},
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
			blockedIPs:    []string{"1.2.3.4", "254.254.254.1"},
			blockedIPNets: []string{"66.67.68.0/28"},
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
			blockedIPs:    []string{"1.2.3.4", "254.254.254.1"},
			blockedIPNets: []string{"66.67.68.0/28"},
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
			additionalBlockedIPs: []net.IP{{254, 254, 254, 1}},
			additionalBlockedIPNets: []*net.IPNet{{
				IP:   net.IP{55, 55, 55, 0},
				Mask: net.IPv4Mask(255, 255, 255, 0),
			}},
			blockedIPs:    []string{"1.2.3.4", "254.254.254.1"},
			blockedIPNets: []string{"66.67.68.0/28", "55.55.55.0/24"},
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			clientCalls := struct {
				m map[models.URL]int
				sync.Mutex
			}{
				m: make(map[models.URL]int),
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

			client := &http.Client{
				Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					url := models.URL(r.URL.String())
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
						body = tc.malicious.content
						err = tc.malicious.clientErr
					case adsBlockListIPsURL:
						body = tc.ads.content
						err = tc.ads.clientErr
					case surveillanceBlockListIPsURL:
						body = tc.surveillance.content
						err = tc.surveillance.clientErr
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

			blockedIPs, blockedIPNets, errs := builder.IPs(ctx,
				tc.malicious.blocked, tc.ads.blocked, tc.surveillance.blocked,
				tc.additionalBlockedIPs, tc.additionalBlockedIPNets)

			assert.ElementsMatch(t, tc.blockedIPs, convertIPsToString(blockedIPs))
			assert.ElementsMatch(t, tc.blockedIPNets, convertIPNetsToString(blockedIPNets))
			assert.ElementsMatch(t, tc.errsString, convertErrorsToString(errs))
		})
	}
}
