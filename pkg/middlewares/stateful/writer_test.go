package stateful

import (
	"errors"
	"net"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testWriter struct {
	t                *testing.T
	expectedResponse *dns.Msg
	err              error
	// to have methods other than WriteMsg we don't use in our tests
	dns.ResponseWriter
}

func (w *testWriter) WriteMsg(response *dns.Msg) error {
	assert.Equal(w.t, w.expectedResponse, response)
	return w.err
}

func Test_Writer(t *testing.T) {
	t.Parallel()

	var dummyResponse = &dns.Msg{Answer: []dns.RR{
		&dns.A{A: net.IP{1, 2, 3, 4}},
	}}
	var errDummy = errors.New("dummy")

	testCases := map[string]struct {
		response *dns.Msg
		err      error
	}{
		"nil response and nil error": {},
		"response and error": {
			response: dummyResponse,
			err:      errDummy,
		},
		"response and no error": {
			response: dummyResponse,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			w := &testWriter{
				t:                t,
				expectedResponse: testCase.response,
				err:              testCase.err,
			}

			writer := NewWriter(w)

			err := writer.WriteMsg(testCase.response)

			if testCase.err != nil {
				require.Error(t, err)
				assert.Equal(t, testCase.err.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
