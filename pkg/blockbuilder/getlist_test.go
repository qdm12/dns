package blockbuilder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getList(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		content   []byte
		status    int
		clientErr error
		results   []string
		err       error
	}{
		"no result": {
			status: http.StatusOK,
		},
		"bad status": {
			status: http.StatusInternalServerError,
			err:    fmt.Errorf("bad HTTP status code: 500 Internal Server Error"),
		},
		"network error": {
			status:    http.StatusOK,
			clientErr: fmt.Errorf("error"),
			err:       fmt.Errorf(`Get "http://irrelevant_url": error`),
		},
		"results": {
			content: []byte("a\nb\nc\n"),
			status:  http.StatusOK,
			results: []string{"a", "b", "c"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			const url = "http://irrelevant_url"

			client := &http.Client{
				Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					assert.Equal(t, url, r.URL.String())
					if tc.clientErr != nil {
						return nil, tc.clientErr
					}
					return &http.Response{
						StatusCode: tc.status,
						Status:     http.StatusText(tc.status),
						Body:       io.NopCloser(bytes.NewReader(tc.content)),
					}, nil
				}),
			}

			results, err := getList(ctx, client, url)
			if tc.err != nil {
				require.Error(t, err)
				assert.Equal(t, tc.err.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.results, results)
		})
	}
}
