package exchanger

import (
	"context"
	"fmt"
	"net"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/mockhelp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Exchanger_exchangeWithPool(t *testing.T) {
	t.Parallel()

	const testDomain = "example.com."
	testIP := net.IP{93, 184, 216, 34}

	const network = "tcp"
	request := new(dns.Msg).SetQuestion(testDomain, dns.TypeA)
	answers := []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{
				Name:     testDomain,
				Rrtype:   dns.TypeA,
				Rdlength: uint16(len(testIP)), //nolint:gosec
			},
			A: testIP,
		},
	}

	localhostString := mockhelp.NewMatcherRegex(`127\.0\.0\.1:[0-9]+`)

	testCases := map[string]struct {
		makeExchanger   func(*testing.T) (e *Exchanger, runError <-chan error)
		makeCtx         func() context.Context
		expectedErr     string
		expectedAnswers []dns.RR
	}{
		"context_canceled": {
			makeExchanger: func(t *testing.T) (e *Exchanger, runError <-chan error) {
				t.Helper()
				dialer, _, runError := startLocalTCPDNS(t, nil)
				ctrl := gomock.NewController(t)
				poolMetrics := NewMockPoolMetrics(ctrl)
				poolMetrics.EXPECT().NewConnsInc(localhostString, "error")
				poolMetrics.EXPECT().GetConnInc(localhostString, "error")
				exchanger := New(dialer, poolMetrics, nil)
				return exchanger, runError
			},
			makeCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expectedErr: "getting test-dialer-tcp-[0-9]+ connection for " +
				"request IN A example.com.: creating connection: " +
				`dial tcp 127\.0\.0\.1:[0-9]+: operation was canceled`,
		},
		"server_closes_connection_first_time_after_write": {
			makeExchanger: func(t *testing.T) (e *Exchanger, runError <-chan error) {
				t.Helper()
				handler := &testHandler{
					sequence: []dns.HandlerFunc{
						dns.HandlerFunc(func(writer dns.ResponseWriter, _ *dns.Msg) {
							_ = writer.Close()
						}),
						dns.HandlerFunc(func(writer dns.ResponseWriter, request *dns.Msg) {
							response := new(dns.Msg)
							response.SetReply(request)
							response.Answer = answers
							_ = writer.WriteMsg(response)
						}),
					},
				}
				dialer, _, runError := startLocalTCPDNS(t, handler)

				ctrl := gomock.NewController(t)
				poolMetrics := NewMockPoolMetrics(ctrl)
				// pool Get call for first connection
				poolMetrics.EXPECT().NewConnsInc(localhostString, "success")
				poolMetrics.EXPECT().GetConnInc(localhostString, "success")
				poolMetrics.EXPECT().LiveConnInc(localhostString)
				// pool Renew call after connection closed by server
				poolMetrics.EXPECT().RenewConnInc(localhostString, "connection error", "success")
				poolMetrics.EXPECT().NewConnsInc(localhostString, "success")
				poolMetrics.EXPECT().RecordLifetime(localhostString, gomock.Any())
				// pool Put call after successful second exchange
				poolMetrics.EXPECT().RecordUseTime(localhostString, gomock.Any())
				poolMetrics.EXPECT().PutConnInc(localhostString, "live")

				exchanger := New(dialer, poolMetrics, nil)
				return exchanger, runError
			},
			expectedAnswers: answers,
		},
		"server_is_stopped": {
			makeExchanger: func(t *testing.T) (e *Exchanger, runError <-chan error) {
				t.Helper()
				handler := &testHandler{}
				dialer, stopServer, runError := startLocalTCPDNS(t, handler)
				stopServer()
				ctrl := gomock.NewController(t)
				poolMetrics := NewMockPoolMetrics(ctrl)
				poolMetrics.EXPECT().NewConnsInc(localhostString, "error")
				poolMetrics.EXPECT().GetConnInc(localhostString, "error")
				exchanger := New(dialer, poolMetrics, nil)
				return exchanger, runError
			},
			expectedErr: `getting test-dialer-tcp-[0-9]+ connection for request IN A example.com.: ` +
				`creating connection: dial tcp 127\.0\.0\.1:[0-9]+: connect: connection refused`,
		},
		"success": {
			makeExchanger: func(t *testing.T) (e *Exchanger, runError <-chan error) {
				t.Helper()
				handler := &testHandler{
					sequence: []dns.HandlerFunc{
						dns.HandlerFunc(func(writer dns.ResponseWriter, request *dns.Msg) {
							response := new(dns.Msg)
							response.SetReply(request)
							response.Answer = answers
							_ = writer.WriteMsg(response)
						}),
					},
				}
				dialer, _, runError := startLocalTCPDNS(t, handler)

				ctrl := gomock.NewController(t)
				poolMetrics := NewMockPoolMetrics(ctrl)
				// pool Get call for first connection
				poolMetrics.EXPECT().NewConnsInc(localhostString, "success")
				poolMetrics.EXPECT().GetConnInc(localhostString, "success")
				poolMetrics.EXPECT().LiveConnInc(localhostString)
				// pool Put call after successful exchange
				poolMetrics.EXPECT().RecordUseTime(localhostString, gomock.Any())
				poolMetrics.EXPECT().PutConnInc(localhostString, "live")

				exchanger := New(dialer, poolMetrics, nil)
				return exchanger, runError
			},
			expectedAnswers: answers,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			exchanger, runError := testCase.makeExchanger(t)

			ctx := context.Background()
			if testCase.makeCtx != nil {
				ctx = testCase.makeCtx()
			}

			response, err := exchanger.exchangeWithPool(ctx, network, request)
			if testCase.expectedErr == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedAnswers, response.Answer)
			} else {
				require.Error(t, err)
				assert.Regexp(t, testCase.expectedErr, err.Error())
				assert.Nil(t, response)
			}

			select {
			case err := <-runError:
				if err != nil {
					t.Errorf("tcp dns server crashed: %s", err)
				}
			default:
			}
		})
	}
}

type testHandler struct {
	i        int
	sequence []dns.HandlerFunc
}

func (h *testHandler) ServeDNS(writer dns.ResponseWriter, request *dns.Msg) {
	if h.i == len(h.sequence) {
		panic("testHandler: no more handlers in sequence")
	}
	handler := h.sequence[h.i]
	h.i++
	handler(writer, request)
}

type testDialer struct {
	reuseConns bool
	port       uint16
}

func (d *testDialer) Dial(ctx context.Context, network, _ string,
) (net.Conn, error) {
	dialer := &net.Dialer{}
	return dialer.DialContext(ctx, network, net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", d.port)))
}

func (d *testDialer) ReusableConnsSupported() bool {
	return d.reuseConns
}

func (d *testDialer) Addresses() []string {
	return []string{
		fmt.Sprintf("127.0.0.1:%d", d.port),
	}
}

func (d *testDialer) String() string {
	return "test-dialer-tcp-" + fmt.Sprint(d.port)
}

func startLocalTCPDNS(t *testing.T, handler dns.Handler,
) (dialer *testDialer, stop func(), runError <-chan error) {
	t.Helper()

	listener, err := net.ListenTCP("tcp", nil)
	require.NoError(t, err)

	server := dns.Server{
		Listener: listener,
		Handler:  handler,
	}

	readyCh := make(chan struct{})
	server.NotifyStartedFunc = func() {
		close(readyCh)
	}

	runErrorCh := make(chan error)
	go func() {
		runErrorCh <- server.ActivateAndServe()
	}()
	isStopped := false
	stop = func() {
		if isStopped {
			return
		}
		isStopped = true
		err := server.Shutdown()
		require.NoError(t, err)
	}
	t.Cleanup(stop)

	select {
	case <-readyCh:
	case err := <-runErrorCh:
		t.Fatal("server failed to start:", err)
	}

	return &testDialer{
		reuseConns: true,
		port:       uint16(listener.Addr().(*net.TCPAddr).Port), //nolint:gosec,forcetypeassert
	}, stop, runErrorCh
}
