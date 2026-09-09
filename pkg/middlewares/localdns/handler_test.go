package localdns

import (
	"context"
	"errors"
	"net/netip"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

type testBlockingExchanger struct {
	once          sync.Once
	started       chan struct{}
	release       chan struct{}
	response      *dns.Msg
	exchangeCalls atomic.Int32
}

func (b *testBlockingExchanger) Exchange(_ context.Context, _ string,
	request *dns.Msg,
) (*dns.Msg, error) {
	b.exchangeCalls.Add(1)
	b.once.Do(func() { close(b.started) })
	<-b.release
	responseCopy := b.response.Copy()
	responseCopy.SetReply(request)
	return responseCopy, nil
}

type testRecordingWriter struct {
	dns.ResponseWriter
	messages []*dns.Msg
}

func (w *testRecordingWriter) WriteMsg(m *dns.Msg) error {
	w.messages = append(w.messages, m)
	return nil
}

func Test_handler_ServeDNS_SingleFlight(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	logger := NewMockLogger(ctrl)

	localResponse := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Rcode: dns.RcodeSuccess,
		},
		Answer: []dns.RR{
			&dns.TXT{Txt: []string{"handled_by_local"}},
		},
	}
	blockingExchanger := &testBlockingExchanger{
		started:  make(chan struct{}),
		release:  make(chan struct{}),
		response: localResponse,
	}

	waiterAttached := make(chan struct{})
	handler := &handler{
		inFlight:       make(map[string]*inFlightExchange),
		localChecker:   local.New(nil),
		logger:         logger,
		next:           dns.HandlerFunc(func(_ dns.ResponseWriter, _ *dns.Msg) {}),
		ctx:            context.Background(),
		localExchanges: []exchangerIntf{blockingExchanger},
		localResolvers: []string{"10.0.0.1:53"},
		waitInFlightNotifier: func() {
			close(waiterAttached)
		},
	}

	makeRequest := func(messageID uint16) *dns.Msg {
		return &dns.Msg{
			MsgHdr: dns.MsgHdr{Id: messageID},
			Question: []dns.Question{{
				Name:   "domain.local.",
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			}},
		}
	}

	leaderWriter := &testRecordingWriter{}
	leaderDone := make(chan struct{})
	go func() {
		handler.ServeDNS(leaderWriter, makeRequest(1))
		close(leaderDone)
	}()

	select {
	case <-blockingExchanger.started:
	case <-time.After(2 * time.Second):
		t.Fatal("exchange did not start")
	}

	waiterWriter := &testRecordingWriter{}
	waiterDone := make(chan struct{})
	go func() {
		handler.ServeDNS(waiterWriter, makeRequest(2))
		close(waiterDone)
	}()

	select {
	case <-waiterAttached:
	case <-time.After(2 * time.Second):
		t.Fatal("waiter did not attach to the in-flight exchange")
	}

	close(blockingExchanger.release)

	for name, done := range map[string]<-chan struct{}{
		"leader": leaderDone,
		"waiter": waiterDone,
	} {
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatalf("%s did not complete", name)
		}
	}

	// Only one exchange is performed for the two concurrent requests
	assert.EqualValues(t, 1, blockingExchanger.exchangeCalls.Load())

	require.Len(t, leaderWriter.messages, 1)
	require.Len(t, waiterWriter.messages, 1)
	for _, writer := range []*testRecordingWriter{leaderWriter, waiterWriter} {
		response := writer.messages[0]
		assert.Equal(t, dns.RcodeSuccess, response.Rcode)
		require.Len(t, response.Answer, 1)
		txt, ok := response.Answer[0].(*dns.TXT)
		require.True(t, ok)
		require.Len(t, txt.Txt, 1)
		assert.Equal(t, "handled_by_local", txt.Txt[0])
	}

	assert.EqualValues(t, 1, leaderWriter.messages[0].Id)
	assert.EqualValues(t, 2, waiterWriter.messages[0].Id)
}

func Test_handler(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	logger := NewMockLogger(ctrl)

	handlerA := dns.HandlerFunc(func(
		writer dns.ResponseWriter, request *dns.Msg,
	) {
		response := new(dns.Msg)
		response.SetRcode(request, dns.RcodeNameError)
		err := writer.WriteMsg(response)
		require.NoError(t, err)
	})
	localAddressA, runErrorA := runLocalDNS(t, handlerA)

	handlerB := dns.HandlerFunc(func(
		writer dns.ResponseWriter, request *dns.Msg,
	) {
		response := new(dns.Msg)
		response.SetReply(request)
		response.Answer = []dns.RR{
			&dns.TXT{
				Hdr: dns.RR_Header{
					Name:   "domain.local.",
					Rrtype: dns.TypeTXT,
				},
				Txt: []string{"B"},
			},
		}

		err := writer.WriteMsg(response)
		require.NoError(t, err)
	})
	localAddressB, runErrorB := runLocalDNS(t, handlerB)

	resolvers := []netip.AddrPort{
		netip.MustParseAddrPort(localAddressA),
		netip.MustParseAddrPort(localAddressB),
	}
	next := dns.HandlerFunc(func(writer dns.ResponseWriter, _ *dns.Msg) {
		response := &dns.Msg{
			Answer: []dns.RR{
				&dns.TXT{Txt: []string{"handled_by_next"}},
			},
		}
		_ = writer.WriteMsg(response)
	})

	localChecker := NewMockLocalChecker(ctrl)
	localChecker.EXPECT().IsFQDNLocal("domain.com.").Return(false)
	localChecker.EXPECT().IsFQDNLocal("domain.local.").Return(true)

	const timeoutWarn = false
	handler := newHandler(resolvers, localChecker, logger, next, timeoutWarn)

	writer := NewMockResponseWriter(ctrl)

	// Public name request goes to next handler
	request := &dns.Msg{
		Question: []dns.Question{{
			Name: "domain.com.", Qtype: dns.TypeTXT,
		}},
	}
	writer.EXPECT().WriteMsg(&dns.Msg{
		Answer: []dns.RR{
			&dns.TXT{Txt: []string{"handled_by_next"}},
		},
	}).Return(nil)
	handler.ServeDNS(writer, request)

	// Local name request goes to local resolvers
	request = &dns.Msg{
		Question: []dns.Question{{
			Name: "domain.local.", Qtype: dns.TypeTXT,
		}},
	}
	expectedFinalResponse := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Response: true,
			Rcode:    dns.RcodeSuccess,
		},
		Question: []dns.Question{{
			Name: "domain.local.", Qtype: dns.TypeTXT,
		}},
		Answer: []dns.RR{
			&dns.TXT{
				Hdr: dns.RR_Header{
					Name:     "domain.local.",
					Rrtype:   dns.TypeTXT,
					Rdlength: 2, // computed by DNS library when writing
				},
				Txt: []string{"B"},
			},
		},
	}
	writer.EXPECT().WriteMsg(expectedFinalResponse).Return(nil)
	handler.ServeDNS(writer, request)

	handler.stop()

	// Check local DNS servers did not crash
	select {
	case err := <-runErrorA:
		t.Error(err)
	case err := <-runErrorB:
		t.Error(err)
	default:
	}
}

func Test_handler_ServeDNS(t *testing.T) {
	t.Parallel()

	errTest := errors.New("test error")

	nextResponse := &dns.Msg{
		Answer: []dns.RR{
			&dns.TXT{Txt: []string{"handled_by_next"}},
		},
	}

	next := dns.HandlerFunc(func(writer dns.ResponseWriter, _ *dns.Msg) {
		_ = writer.WriteMsg(nextResponse)
	})

	makeTestExchanger := func(ctrl *gomock.Controller,
		ctx context.Context, response *dns.Msg, err error,
	) *MockexchangerIntf {
		exchanger := NewMockexchangerIntf(ctrl)
		exchanger.EXPECT().Exchange(ctx, "udp",
			gomock.AssignableToTypeOf(&dns.Msg{})).
			Return(response, err)
		return exchanger
	}

	testCases := map[string]struct {
		request     *dns.Msg
		makeHandler func(ctrl *gomock.Controller) *handler
		response    *dns.Msg
	}{
		"no_question": {
			request: &dns.Msg{},
			makeHandler: func(_ *gomock.Controller) *handler {
				return &handler{
					next: next,
				}
			},
			response: nextResponse,
		},
		"multiple_questions": {
			request: &dns.Msg{
				Question: []dns.Question{{}, {}},
			},
			makeHandler: func(_ *gomock.Controller) *handler {
				return &handler{
					next: next,
				}
			},
			response: nextResponse,
		},
		"public_name": {
			request: &dns.Msg{
				Question: []dns.Question{{
					Name: "domain.com.",
				}},
			},

			makeHandler: func(_ *gomock.Controller) *handler {
				return &handler{
					localChecker: local.New(nil),
					next:         next,
				}
			},
			response: nextResponse,
		},
		"local_name_exchange_error": {
			request: &dns.Msg{
				Question: []dns.Question{{
					Qclass: dns.ClassINET,
					Qtype:  dns.TypeA,
					Name:   "domain.local.",
				}},
			},
			makeHandler: func(ctrl *gomock.Controller) *handler {
				logger := NewMockLogger(ctrl)
				logger.EXPECT().Warn("exchanging over udp: test error")

				ctx := context.Background()

				localExchanges := []exchangerIntf{
					makeTestExchanger(ctrl, ctx, nil, errTest),
				}

				return &handler{
					localChecker:   local.New(nil),
					ctx:            ctx,
					logger:         logger,
					next:           next,
					localExchanges: localExchanges,
					localResolvers: []string{"10.0.0.1:53"},
					inFlight:       make(map[string]*inFlightExchange),
				}
			},
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Response: true,
					Rcode:    dns.RcodeNameError,
				},
				Question: []dns.Question{{
					Qclass: dns.ClassINET,
					Qtype:  dns.TypeA,
					Name:   "domain.local.",
				}},
			},
		},
		"local_name_failure_rcode": {
			request: &dns.Msg{
				Question: []dns.Question{{
					Name: "domain.local.",
				}},
			},
			makeHandler: func(ctrl *gomock.Controller) *handler {
				logger := NewMockLogger(ctrl)
				logger.EXPECT().Debug("response received for " +
					"domain.local. from 10.0.0.1:53 over udp has " +
					"rcode REFUSED")

				ctx := context.Background()

				localExchanges := []exchangerIntf{
					makeTestExchanger(ctrl, ctx, &dns.Msg{
						MsgHdr: dns.MsgHdr{
							Rcode: dns.RcodeRefused,
						},
					}, nil),
				}

				return &handler{
					ctx:            ctx,
					logger:         logger,
					localChecker:   local.New(nil),
					next:           next,
					localExchanges: localExchanges,
					localResolvers: []string{"10.0.0.1:53"},
					inFlight:       make(map[string]*inFlightExchange),
				}
			},
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Response: true,
					Rcode:    dns.RcodeNameError,
				},
				Question: []dns.Question{{
					Name: "domain.local.",
				}},
			},
		},
		"local_name_success": {
			request: &dns.Msg{
				Question: []dns.Question{{
					Name: "domain.local.",
				}},
			},
			makeHandler: func(ctrl *gomock.Controller) *handler {
				ctx := context.Background()

				localExchanges := []exchangerIntf{
					makeTestExchanger(ctrl, ctx, &dns.Msg{
						MsgHdr: dns.MsgHdr{
							Rcode: dns.RcodeSuccess,
						},
						Answer: []dns.RR{
							&dns.TXT{Txt: []string{"handled_by_local"}},
						},
					}, nil),
				}

				return &handler{
					localChecker:   local.New(nil),
					ctx:            ctx,
					next:           next,
					localExchanges: localExchanges,
					localResolvers: []string{"10.0.0.1:53"},
					inFlight:       make(map[string]*inFlightExchange),
				}
			},
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Rcode: dns.RcodeSuccess,
				},
				Answer: []dns.RR{
					&dns.TXT{Txt: []string{"handled_by_local"}},
				},
			},
		},
		"local_name_success_after_failures": {
			request: &dns.Msg{
				Question: []dns.Question{{
					Qclass: dns.ClassINET,
					Qtype:  dns.TypeA,
					Name:   "domain.local.",
				}},
			},
			makeHandler: func(ctrl *gomock.Controller) *handler {
				ctx := context.Background()

				localExchanges := []exchangerIntf{
					makeTestExchanger(ctrl, ctx, nil, errTest), // exchange error
					makeTestExchanger(ctrl, ctx, &dns.Msg{
						MsgHdr: dns.MsgHdr{
							Rcode: dns.RcodeRefused,
						},
					}, nil), // rcode not success
					makeTestExchanger(ctrl, ctx, &dns.Msg{
						MsgHdr: dns.MsgHdr{
							Rcode: dns.RcodeSuccess,
						},
						Answer: []dns.RR{
							&dns.TXT{Txt: []string{"handled_by_local"}},
						},
					}, nil), // success
					NewMockexchangerIntf(ctrl), // unused
				}

				logger := NewMockLogger(ctrl)
				logger.EXPECT().Warn("exchanging over udp: test error")
				logger.EXPECT().Debug("response received for " +
					"domain.local. from 10.0.0.2:53 over udp has " +
					"rcode REFUSED")

				return &handler{
					localChecker:   local.New(nil),
					ctx:            ctx,
					logger:         logger,
					next:           next,
					localExchanges: localExchanges,
					localResolvers: []string{
						"10.0.0.1:53", "10.0.0.2:53",
						"10.0.0.3:53", "10.0.0.4:53",
					},
					inFlight: make(map[string]*inFlightExchange),
				}
			},
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Rcode: dns.RcodeSuccess,
				},
				Answer: []dns.RR{
					&dns.TXT{Txt: []string{"handled_by_local"}},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			handler := testCase.makeHandler(ctrl)

			writer := NewMockResponseWriter(ctrl)
			writer.EXPECT().WriteMsg(testCase.response)

			handler.ServeDNS(writer, testCase.request)
		})
	}
}
