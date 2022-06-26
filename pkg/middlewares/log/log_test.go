package log

import (
	"errors"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	t.Parallel()

	request := &dns.Msg{Question: []dns.Question{
		{Name: "question"},
	}}
	remoteAddress := &net.UDPAddr{
		IP:   net.IP{1, 2, 3, 4},
		Port: 8000,
	}

	ctrl := gomock.NewController(t)
	logger := NewMockLogger(ctrl)
	logger.EXPECT().Log(remoteAddress, request, nil)

	settings := Settings{
		Logger: logger,
	}

	middleware := New(settings)

	next := dns.HandlerFunc(func(rw dns.ResponseWriter, m *dns.Msg) {})
	handler := middleware(next)

	writer := NewMockResponseWriter(ctrl)
	writer.EXPECT().RemoteAddr().Return(remoteAddress)

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

	request := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id: 1,
		},
		Question: []dns.Question{
			{Name: "question"},
		},
	}
	response := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id: 1,
		},
		Answer: []dns.RR{
			&dns.A{A: net.IP{1, 2, 3, 4}},
		},
	}

	testCases := map[string]struct {
		handlerErr error
	}{
		"handler error": {
			handlerErr: errors.New("dummy"),
		},
		"success": {},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			logger := NewMockLogger(ctrl)
			remoteAddress := &net.UDPAddr{
				IP:   net.IP{1, 2, 3, 4},
				Port: 8000,
			}
			logger.EXPECT().Log(remoteAddress, request, response)

			if testCase.handlerErr != nil {
				logger.EXPECT().Error(request.Id,
					"cannot write DNS response: "+testCase.handlerErr.Error())
			}

			next := dns.HandlerFunc(func(rw dns.ResponseWriter, m *dns.Msg) {
				assert.Equal(t, request, m)
				_ = rw.WriteMsg(response)
			})

			handler := &handler{
				logger: logger,
				next:   next,
			}

			responseWriter := NewMockResponseWriter(ctrl)
			responseWriter.EXPECT().RemoteAddr().Return(remoteAddress)
			writer := &testWriter{
				err:            testCase.handlerErr,
				ResponseWriter: responseWriter,
			}
			handler.ServeDNS(writer, request)
		})
	}
}
