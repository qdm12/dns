package provider

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Parse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		s        string
		provider Provider
		err      error
	}{
		"empty string": {
			err: errors.New(`cannot parse provider: ""`),
		},
		"bad provider string": {
			s:   "invalid",
			err: errors.New(`cannot parse provider: "invalid"`),
		},
		"cirafamily": {
			s:        "cirafamily",
			provider: CiraFamily(),
		},
		"ciraprivate": {
			s:        "ciraprivate",
			provider: CiraPrivate(),
		},
		"ciraprotected": {
			s:        "ciraprotected",
			provider: CiraProtected(),
		},
		"cleanbrowsingadult": {
			s:        "cleanbrowsingadult",
			provider: CleanBrowsingAdult(),
		},
		"cleanbrowsingfamily": {
			s:        "cleanbrowsingfamily",
			provider: CleanBrowsingFamily(),
		},
		"cleanbrowsingsecurity": {
			s:        "cleanbrowsingsecurity",
			provider: CleanBrowsingSecurity(),
		},
		"cloudflare": {
			s:        "cloudflare",
			provider: Cloudflare(),
		},
		"cloudflarefamily": {
			s:        "cloudflarefamily",
			provider: CloudflareFamily(),
		},
		"cloudflaresecurity": {
			s:        "cloudflaresecurity",
			provider: CloudflareSecurity(),
		},
		"google": {
			s:        "google",
			provider: Google(),
		},
		"libredns": {
			s:        "libredns",
			provider: LibreDNS(),
		},
		"quad9": {
			s:        "quad9",
			provider: Quad9(),
		},
		"quad9secured": {
			s:        "quad9secured",
			provider: Quad9Secured(),
		},
		"quad9unsecured": {
			s:        "quad9unsecured",
			provider: Quad9Unsecured(),
		},
		"quadrant": {
			s:        "quadrant",
			provider: Quadrant(),
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			provider, err := Parse(testCase.s)

			if testCase.err != nil {
				require.Error(t, err)
				assert.Equal(t, testCase.err.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.provider, provider)
		})
	}
}
