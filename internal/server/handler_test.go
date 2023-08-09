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
	filter := NewMockFilter(nil)
	cache := NewMockCache(nil)
	logger := NewMockLogger(nil)

	handler := New(ctx, nil, filter, cache, logger)

	expectedHandler := &Handler{
		ctx:      ctx,
		exchange: nil, // cannot compare functions
		filter:   filter,
		cache:    cache,
		logger:   logger,
	}
	assert.Equal(t, expectedHandler, handler)
}

func Test_Handler_ServeDNS(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		makeHandler func(ctrl *gomock.Controller) *Handler
		request     *dns.Msg
		response    *dns.Msg
	}{
		"filtered_request": {
			makeHandler: func(ctrl *gomock.Controller) *Handler {
				filter := NewMockFilter(ctrl)
				filter.EXPECT().FilterRequest(&dns.Msg{
					Question: []dns.Question{{Name: "test"}}},
				).Return(true)
				return &Handler{
					filter: filter,
				}
			},
			request: &dns.Msg{
				Question: []dns.Question{{Name: "test"}},
			},
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Response: true,
					Rcode:    dns.RcodeRefused,
				},
				Question: []dns.Question{{Name: "test"}},
			},
		},
		"filtered_cached_response": {
			makeHandler: func(ctrl *gomock.Controller) *Handler {
				expectedRequest := &dns.Msg{
					Question: []dns.Question{{Name: "test"}},
				}

				filter := NewMockFilter(ctrl)
				filter.EXPECT().FilterRequest(expectedRequest).Return(false)

				cache := NewMockCache(ctrl)
				cachedResponse := &dns.Msg{Answer: []dns.RR{&dns.A{}}}
				cache.EXPECT().Get(expectedRequest).Return(cachedResponse)

				filter.EXPECT().FilterResponse(cachedResponse).Return(true)

				cache.EXPECT().Remove(expectedRequest)

				return &Handler{
					filter: filter,
					cache:  cache,
				}
			},
			request: &dns.Msg{
				Question: []dns.Question{{Name: "test"}},
			},
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Response: true,
					Rcode:    dns.RcodeRefused,
				},
				Question: []dns.Question{{Name: "test"}},
			},
		},
		"cached_response": {
			makeHandler: func(ctrl *gomock.Controller) *Handler {
				expectedRequest := &dns.Msg{
					Question: []dns.Question{{Name: "test"}},
				}

				filter := NewMockFilter(ctrl)
				filter.EXPECT().FilterRequest(expectedRequest).Return(false)

				cache := NewMockCache(ctrl)
				cachedResponse := &dns.Msg{Answer: []dns.RR{&dns.A{}}}
				cache.EXPECT().Get(expectedRequest).Return(cachedResponse)

				filter.EXPECT().FilterResponse(cachedResponse).Return(false)

				return &Handler{
					filter: filter,
					cache:  cache,
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
		"exchange_error": {
			makeHandler: func(ctrl *gomock.Controller) *Handler {
				expectedRequest := &dns.Msg{
					Question: []dns.Question{{Name: "test"}},
				}

				filter := NewMockFilter(ctrl)
				filter.EXPECT().FilterRequest(expectedRequest).Return(false)

				cache := NewMockCache(ctrl)
				cache.EXPECT().Get(expectedRequest).Return(nil)

				exchange := func(ctx context.Context, request *dns.Msg) (
					response *dns.Msg, err error,
				) {
					return nil, errors.New("test error")
				}

				logger := NewMockLogger(ctrl)
				logger.EXPECT().Warn("test error")

				return &Handler{
					filter:   filter,
					cache:    cache,
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
		"filtered_exchanged_response": {
			makeHandler: func(ctrl *gomock.Controller) *Handler {
				expectedRequest := &dns.Msg{
					Question: []dns.Question{{Name: "test"}},
				}

				filter := NewMockFilter(ctrl)
				filter.EXPECT().FilterRequest(expectedRequest).Return(false)

				cache := NewMockCache(ctrl)
				cache.EXPECT().Get(expectedRequest).Return(nil)

				expectedResponse := &dns.Msg{Answer: []dns.RR{&dns.A{}}}

				exchange := func(ctx context.Context, request *dns.Msg) (
					response *dns.Msg, err error,
				) {
					return expectedResponse, nil
				}

				filter.EXPECT().
					FilterResponse(expectedResponse).
					Return(true)

				return &Handler{
					filter:   filter,
					cache:    cache,
					exchange: exchange,
				}
			},
			request: &dns.Msg{
				Question: []dns.Question{{Name: "test"}},
			},
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Response: true,
					Rcode:    dns.RcodeRefused,
				},
				Question: []dns.Question{{Name: "test"}},
			},
		},
		"exchanged_response": {
			makeHandler: func(ctrl *gomock.Controller) *Handler {
				expectedRequest := &dns.Msg{
					Question: []dns.Question{{Name: "test"}},
				}

				filter := NewMockFilter(ctrl)
				filter.EXPECT().FilterRequest(expectedRequest).Return(false)

				cache := NewMockCache(ctrl)
				cache.EXPECT().Get(expectedRequest).Return(nil)

				expectedResponse := &dns.Msg{Answer: []dns.RR{&dns.A{}}}

				exchange := func(ctx context.Context, request *dns.Msg) (
					response *dns.Msg, err error,
				) {
					return expectedResponse, nil
				}

				filter.EXPECT().
					FilterResponse(expectedResponse).
					Return(false)

				cache.EXPECT().Add(expectedRequest, expectedResponse)

				return &Handler{
					filter:   filter,
					cache:    cache,
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
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			handler := testCase.makeHandler(ctrl)
			writer := &testWriter{}

			handler.ServeDNS(writer, testCase.request)

			assert.Equal(t, testCase.response, writer.responseWritten)
		})
	}
}
