package blockbuilder

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Builder_Hostnames(t *testing.T) {
	t.Parallel()
	type blockParams struct {
		blocked   bool
		content   []byte
		clientErr error
	}
	tests := map[string]struct {
		malicious                  blockParams
		ads                        blockParams
		surveillance               blockParams
		additionalBlockedHostnames []string
		additionalAllowedHostnames []string
		blockedHostnames           []string
		errsString                 []string
	}{
		"nothing blocked": {},
		"only malicious blocked": {
			malicious: blockParams{
				blocked: true,
				content: []byte("site_a\nsite_b"),
			},
			blockedHostnames: []string{"site_a", "site_b"},
		},
		"all blocked with some duplicates": {
			malicious: blockParams{
				blocked: true,
				content: []byte("site_a\nsite_b"),
			},
			ads: blockParams{
				blocked: true,
				content: []byte("site_a\nsite_c"),
			},
			surveillance: blockParams{
				blocked: true,
				content: []byte("site_c\nsite_a"),
			},
			blockedHostnames: []string{"site_a", "site_b", "site_c"},
			errsString:       nil,
		},
		"all blocked with one errored": {
			malicious: blockParams{
				blocked: true,
				content: []byte("site_a\nsite_b"),
			},
			ads: blockParams{
				blocked: true,
				content: []byte("site_a\nsite_c"),
			},
			surveillance: blockParams{
				blocked:   true,
				clientErr: fmt.Errorf("surveillance error"),
			},
			blockedHostnames: []string{"site_a", "site_b", "site_c"},
			errsString: []string{
				`Get "https://raw.githubusercontent.com/qdm12/files/master/surveillance-hostnames.updated": surveillance error`,
			},
		},
		"blocked with allowed hostnames": {
			malicious: blockParams{
				blocked: true,
				content: []byte("site_a\nsite_b"),
			},
			ads: blockParams{
				blocked: true,
				content: []byte("site_c\nsite_d"),
			},
			additionalAllowedHostnames: []string{"site_b", "site_c"},
			blockedHostnames:           []string{"site_a", "site_d"},
		},
		"blocked with additional blocked hostnames": {
			malicious: blockParams{
				blocked: true,
				content: []byte("site_a\nsite_b"),
			},
			ads: blockParams{
				blocked: true,
				content: []byte("site_c\nsite_d"),
			},
			additionalAllowedHostnames: []string{"site_b", "site_c"},
			additionalBlockedHostnames: []string{"site_e", "site_b"},
			blockedHostnames:           []string{"site_a", "site_d", "site_e"},
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
				clientCalls.m[maliciousBlockListHostnamesURL] = 0
			}
			if tc.ads.blocked {
				clientCalls.m[adsBlockListHostnamesURL] = 0
			}
			if tc.surveillance.blocked {
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
					case maliciousBlockListHostnamesURL:
						body = tc.malicious.content
						err = tc.malicious.clientErr
					case adsBlockListHostnamesURL:
						body = tc.ads.content
						err = tc.ads.clientErr
					case surveillanceBlockListHostnamesURL:
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

			settings := Settings{Client: client}
			builder := New(settings)

			blockedHostnames, errs := builder.buildHostnames(ctx,
				tc.malicious.blocked, tc.ads.blocked, tc.surveillance.blocked,
				tc.additionalBlockedHostnames, tc.additionalAllowedHostnames)
			var errsString []string
			for _, err := range errs {
				errsString = append(errsString, err.Error())
			}
			assert.ElementsMatch(t, tc.errsString, errsString)
			assert.ElementsMatch(t, tc.blockedHostnames, blockedHostnames)
		})
	}
}
