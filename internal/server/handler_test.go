package server

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := NewMockLogger(nil)

	handler := New(ctx, nil, logger)

	expectedHandler := &Handler{
		ctx:      ctx,
		exchange: nil, // cannot compare functions
		logger:   logger,
	}
	assert.Equal(t, expectedHandler, handler)
}

func Test_Handler_ServeDNS(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		makeHandler func(t *testing.T, ctrl *gomock.Controller) *Handler
		request     *dns.Msg
		response    *dns.Msg
	}{
		"exchange_error": {
			makeHandler: func(t *testing.T, ctrl *gomock.Controller) *Handler {
				t.Helper()
				expectedRequest := &dns.Msg{
					Question: []dns.Question{{Name: "test"}},
				}

				exchange := func(_ context.Context, request *dns.Msg) (
					response *dns.Msg, err error,
				) {
					assert.Equal(t, expectedRequest, request)
					return nil, errors.New("test error")
				}

				logger := NewMockLogger(ctrl)
				logger.EXPECT().Warn("test error")

				return &Handler{
					exchange: exchange,
					logger:   logger,
				}
			},
			request: &dns.Msg{
				Question: []dns.Question{{Name: "test"}},
			},
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Response: true,
					Rcode:    dns.RcodeServerFailure,
				},
				Question: []dns.Question{{Name: "test"}},
			},
		},
		"exchanged_response": {
			makeHandler: func(t *testing.T, _ *gomock.Controller) *Handler {
				t.Helper()
				expectedRequest := &dns.Msg{
					Question: []dns.Question{{Name: "test"}},
				}
				exchange := func(_ context.Context, request *dns.Msg) (
					response *dns.Msg, err error,
				) {
					assert.Equal(t, expectedRequest, request)
					return &dns.Msg{Answer: []dns.RR{&dns.A{}}}, nil
				}

				return &Handler{
					exchange: exchange,
				}
			},
			request: &dns.Msg{
				Question: []dns.Question{{Name: "test"}},
			},
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Response: true,
				},
				Question: []dns.Question{{Name: "test"}},
				Answer:   []dns.RR{&dns.A{}},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			handler := testCase.makeHandler(t, ctrl)
			writer := &testWriter{}

			handler.ServeDNS(writer, testCase.request)

			assert.Equal(t, testCase.response, writer.responseWritten)
		})
	}
}
