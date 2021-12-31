package log

import (
	"errors"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

//go:generate mockgen -destination=mock_format_test.go -package $GOPACKAGE -mock_names Interface=MockFormat github.com/qdm12/dns/pkg/middlewares/log/format Interface
//go:generate mockgen -destination=mock_logger_test.go -package $GOPACKAGE -mock_names Interface=MockLogger github.com/qdm12/dns/pkg/middlewares/log/logger Interface

func Test_New(t *testing.T) {
	t.Parallel()

	request := &dns.Msg{Question: []dns.Question{
		{Name: "question"},
	}}

	ctrl := gomock.NewController(t)

	formatter := NewMockFormat(ctrl)
	formatter.EXPECT().Request(request).Return("formatted request")
	formatter.EXPECT().Response(nil).Return("formatted response")
	formatter.EXPECT().RequestResponse(request, nil).Return("formatted request => response")

	logger := NewMockLogger(ctrl)
	logger.EXPECT().LogRequest("formatted request")
	logger.EXPECT().LogResponse("formatted response")
	logger.EXPECT().LogRequestResponse("formatted request => response")

	settings := Settings{
		Formatter: formatter,
		Logger:    logger,
	}

	middleware := New(settings)

	next := dns.HandlerFunc(func(rw dns.ResponseWriter, m *dns.Msg) {})
	handler := middleware(next)

	var writer dns.ResponseWriter // nil as next does not use it
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

			formatter := NewMockFormat(ctrl)
			formatter.EXPECT().Request(request).Return("formatted request")
			formatter.EXPECT().Response(response).Return("formatted response")
			formatter.EXPECT().RequestResponse(request, response).Return("formatted request => response")

			logger := NewMockLogger(ctrl)
			logger.EXPECT().LogRequest("formatted request")
			logger.EXPECT().LogResponse("formatted response")
			logger.EXPECT().LogRequestResponse("formatted request => response")

			if testCase.handlerErr != nil {
				formatter.EXPECT().Error(request.Id,
					"cannot write DNS response: "+testCase.handlerErr.Error()).
					Return("formatted error")
				logger.EXPECT().Error("formatted error")
			}

			next := dns.HandlerFunc(func(rw dns.ResponseWriter, m *dns.Msg) {
				assert.Equal(t, request, m)
				_ = rw.WriteMsg(response)
			})

			handler := &handler{
				formatter: formatter,
				logger:    logger,
				next:      next,
			}

			writer := &testWriter{err: testCase.handlerErr}
			handler.ServeDNS(writer, request)
		})
	}
}
