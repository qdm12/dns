package main

// import (
// 	"bytes"
// 	"context"
// 	"crypto/tls"
// 	"errors"
// 	"fmt"
// 	"io/ioutil"
// 	"log"
// 	"net"
// 	"net/http"
// 	"net/url"
// 	"os"
// 	"os/signal"
// 	"sync"
// 	"syscall"
// 	"time"

// 	"github.com/miekg/dns"
// )

// func main() {
// 	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
// 	defer stop()

// 	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

// 	server, err := NewServer(ctx, logger, Cloudflare())
// 	if err != nil {
// 		logger.Println(err)
// 		return
// 	}

// 	stopped := make(chan struct{})
// 	go server.Run(ctx, stopped)

// 	select {
// 	case <-ctx.Done():
// 	case <-stopped: // server crashed
// 	}
// 	stop() // stop catching OS signals to exit when receiving an OS signal
// 	<-stopped
// }

// type Provider struct {
// 	serverIPv4 net.IP
// 	serverIPv6 net.IP
// 	serverName string
// 	dohURL     url.URL
// }

// func Cloudflare() Provider {
// 	return Provider{
// 		serverIPv4: net.IP{1, 1, 1, 1},
// 		serverIPv6: net.IP{0x26, 0x6, 0x47, 0x0, 0x47, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x11, 0x11},
// 		serverName: "cloudflare-dns.com",
// 		dohURL: url.URL{
// 			Scheme: "https",
// 			Host:   "cloudflare-dns.com",
// 			Path:   "/dns-query",
// 		},
// 	}
// }

// type Server interface {
// 	Run(ctx context.Context, stopped chan<- struct{})
// }

// type server struct {
// 	dnsServer dns.Server
// 	logger    *log.Logger
// }

// func NewServer(ctx context.Context, logger *log.Logger, provider Provider) (
// 	s Server, err error) {
// 	handler, err := newDNSHandler(ctx, logger, provider)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &server{
// 		dnsServer: dns.Server{
// 			Addr:    ":53",
// 			Net:     "udp",
// 			Handler: handler,
// 		},
// 		logger: logger,
// 	}, nil
// }

// func (s *server) Run(ctx context.Context, stopped chan<- struct{}) {
// 	defer close(stopped)

// 	go func() { // shutdown goroutine
// 		<-ctx.Done()

// 		const graceTime = 100 * time.Millisecond
// 		ctx, cancel := context.WithTimeout(context.Background(), graceTime)
// 		defer cancel()
// 		if err := s.dnsServer.ShutdownContext(ctx); err != nil {
// 			s.logger.Println("DNS server shutdown error: ", err)
// 		}
// 	}()

// 	s.logger.Println("DNS server listening on :53")
// 	if err := s.dnsServer.ListenAndServe(); err != nil {
// 		s.logger.Println("DNS server crashed: ", err)
// 	}
// 	s.logger.Println("DNS server stopped")
// }

// var ErrNoIPWorking = errors.New("both IPv4 and IPv6 do not work")

// func newDNSHandler(ctx context.Context, logger *log.Logger, provider Provider) (
// 	handler dns.Handler, err error) {
// 	ipv4, ipv6 := ipVersionsSupported(ctx)
// 	if !ipv4 && !ipv6 {
// 		return nil, ErrNoIPWorking
// 	}

// 	serverIP := provider.serverIPv4
// 	if ipv6 {
// 		// use IPv6 address by default
// 		// if both IPv4 and IPv6 are supported.
// 		serverIP = provider.serverIPv6
// 	}

// 	client := newDoTClient(serverIP, provider.serverName)
// 	const httpTimeout = 3 * time.Second
// 	client.Timeout = httpTimeout

// 	httpBufferPool := &sync.Pool{
// 		New: func() interface{} {
// 			return bytes.NewBuffer(nil)
// 		},
// 	}

// 	const udpPacketMaxSize = 512
// 	udpBufferPool := &sync.Pool{
// 		New: func() interface{} {
// 			return make([]byte, udpPacketMaxSize)
// 		},
// 	}

// 	return &dnsHandler{
// 		ctx:            ctx,
// 		provider:       provider,
// 		client:         client,
// 		httpBufferPool: httpBufferPool,
// 		udpBufferPool:  udpBufferPool,
// 		logger:         logger,
// 	}, nil
// }

// type dnsHandler struct {
// 	ctx            context.Context
// 	provider       Provider
// 	client         *http.Client
// 	httpBufferPool *sync.Pool
// 	udpBufferPool  *sync.Pool
// 	logger         *log.Logger
// }

// func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
// 	buffer := h.udpBufferPool.Get().([]byte)
// 	// no need to reset buffer as PackBuffer takes care of slicing it down
// 	wire, err := r.PackBuffer(buffer)
// 	if err != nil {
// 		h.logger.Printf("cannot pack message to wire format: %s\n", err)
// 		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
// 		return
// 	}

// 	respWire, err := h.requestHTTP(h.ctx, wire)

// 	// It's fine to copy the slice headers as long as we keep
// 	// the underlying array of bytes.
// 	h.udpBufferPool.Put(buffer) //nolint:staticcheck

// 	if err != nil {
// 		h.logger.Printf("HTTP request failed: %s\n", err)
// 		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
// 		return
// 	}

// 	message := new(dns.Msg)
// 	if err := message.Unpack(respWire); err != nil {
// 		h.logger.Printf("cannot unpack message from wireformat: %s\n", err)
// 		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
// 		return
// 	}

// 	message.SetReply(r)
// 	if err := w.WriteMsg(message); err != nil {
// 		h.logger.Printf("write dns message error: %s\n", err)
// 	}
// }

// var (
// 	ErrHTTPStatus = errors.New("bad HTTP status")
// )

// func (h *dnsHandler) requestHTTP(ctx context.Context, wire []byte) (respWire []byte, err error) {
// 	buffer := h.httpBufferPool.Get().(*bytes.Buffer)
// 	buffer.Reset()
// 	defer h.httpBufferPool.Put(buffer)
// 	_, err = buffer.Write(wire)
// 	if err != nil {
// 		return nil, err
// 	}

// 	request, err := http.NewRequestWithContext(ctx, http.MethodPost, h.provider.dohURL.String(), buffer)
// 	if err != nil {
// 		return nil, err
// 	}

// 	request.Header.Set("Content-Type", "application/dns-udpwireformat")

// 	response, err := h.client.Do(request)

// 	if err != nil {
// 		return nil, err
// 	}
// 	defer response.Body.Close()

// 	if response.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("%w: %s", ErrHTTPStatus, response.Status)
// 	}

// 	respWire, err = ioutil.ReadAll(response.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err := response.Body.Close(); err != nil {
// 		return nil, err
// 	}

// 	return respWire, nil
// }

// func ipVersionsSupported(ctx context.Context) (ipv4, ipv6 bool) {
// 	dialer := &net.Dialer{}
// 	_, err := dialer.DialContext(ctx, "tcp4", "127.0.0.1:0")
// 	ipv4 = err.Error() == "dial tcp4 127.0.0.1:0: connect: connection refused"
// 	_, err = dialer.DialContext(ctx, "tcp6", "[::1]:0")
// 	ipv6 = err.Error() == "dial tcp6 [::1]:0: connect: connection refused"
// 	return ipv4, ipv6
// }

// func newDoTClient(serverIP net.IP, serverName string) *http.Client {
// 	httpTransport := http.DefaultTransport.(*http.Transport).Clone()
// 	dialer := &net.Dialer{
// 		Resolver: newOpportunisticDoTResolver(serverIP, serverName),
// 	}
// 	httpTransport.DialContext = dialer.DialContext
// 	return &http.Client{
// 		Transport: httpTransport,
// 	}
// }

// func newOpportunisticDoTResolver(serverIP net.IP, serverName string) *net.Resolver {
// 	const dialerTimeout = 5 * time.Second
// 	dialer := &net.Dialer{
// 		Timeout: dialerTimeout,
// 	}

// 	plainAddr := net.JoinHostPort(serverIP.String(), "53")
// 	tlsAddr := net.JoinHostPort(serverIP.String(), "853")

// 	tlsConf := &tls.Config{
// 		MinVersion: tls.VersionTLS12,
// 		ServerName: serverName,
// 	}

// 	return &net.Resolver{
// 		PreferGo:     true,
// 		StrictErrors: true,
// 		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
// 			conn, err := dialer.DialContext(ctx, "tcp", tlsAddr)
// 			if err != nil {
// 				// fallback on plain DNS if DoT does not work
// 				return dialer.DialContext(ctx, "udp", plainAddr)
// 			}
// 			return tls.Client(conn, tlsConf), nil
// 		},
// 	}
// }
