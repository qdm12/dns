package log

import (
	"errors"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/miekg/dns"
	"github.com/qdm12/golibs/logging/mock_logging"
	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	t.Parallel()

	var request = &dns.Msg{Question: []dns.Question{
		{Name: "question"},
	}}

	ctrl := gomock.NewController(t)

	logger := mock_logging.NewMockLogger(ctrl)
	logger.EXPECT().Info([]interface{}{"question CLASS0 None"})
	settings := Settings{LogRequests: true}

	middleware := New(logger, settings)

	next := dns.HandlerFunc(func(rw dns.ResponseWriter, m *dns.Msg) {})
	handler := middleware(next)

	var writer dns.ResponseWriter = nil // nil as next does not use it
	handler.ServeDNS(writer, request)
}

type testWriter struct {
	err error
	// to have methods other than WriteMsg we don't use in our tests
	dns.ResponseWriter
}

func (w *testWriter) WriteMsg(response *dns.Msg) error {
	return w.err
}

func Test_handler_ServeDNS(t *testing.T) {
	t.Parallel()

	strPtr := func(s string) *string { return &s }

	var errDummy = errors.New("dummy")
	var request = &dns.Msg{Question: []dns.Question{
		{Name: "question"},
	}}
	var response = &dns.Msg{Answer: []dns.RR{
		&dns.A{A: net.IP{1, 2, 3, 4}},
	}}

	testCases := map[string]struct {
		writer   *testWriter
		settings Settings
		logErr   *string // nil means no logging
		logInfo  *string // nil means no logging
	}{
		"disabled logger": {
			writer: &testWriter{},
		},
		"disabled logger and error": {
			writer: &testWriter{err: errDummy},
			logErr: strPtr("question CLASS0 None: cannot write DNS response: dummy"),
		},
		"log requests only": {
			writer: &testWriter{},
			settings: Settings{
				LogRequests: true,
			},
			logInfo: strPtr("question CLASS0 None"),
		},
		"log responses only": {
			writer: &testWriter{},
			settings: Settings{
				LogResponses: true,
			},
			logInfo: strPtr("[0 CLASS0 None 1.2.3.4]"),
		},
		"log requests and responses": {
			writer: &testWriter{},
			settings: Settings{
				LogRequests:  true,
				LogResponses: true,
			},
			logInfo: strPtr("question CLASS0 None => [0 CLASS0 None 1.2.3.4]"),
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			logger := mock_logging.NewMockLogger(ctrl)
			if testCase.logErr != nil {
				logger.EXPECT().Error([]interface{}{*testCase.logErr})
			}
			if testCase.logInfo != nil {
				logger.EXPECT().Info([]interface{}{*testCase.logInfo})
			}

			next := dns.HandlerFunc(func(rw dns.ResponseWriter, m *dns.Msg) {
				assert.Equal(t, request, m)
				_ = rw.WriteMsg(response)
			})

			handler := &handler{
				logger:   logger,
				next:     next,
				settings: testCase.settings,
			}

			handler.ServeDNS(testCase.writer, request)
		})
	}
}
