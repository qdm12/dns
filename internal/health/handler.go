package health

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

func newHandler(ctx context.Context, dnsListenAddr string) http.Handler {
	return &handler{
		ctx:           ctx,
		dnsListenAddr: dnsListenAddr,
	}
}

type handler struct {
	ctx           context.Context //nolint:containedctx
	dnsListenAddr string
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || (r.RequestURI != "" && r.RequestURI != "/") {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	err := h.isHealthy(h.ctx) //nolint:contextcheck
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// isHealthy checks the localhost DNS server is working by
// resolving github.com.
func (h *handler) isHealthy(ctx context.Context) (err error) {
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", h.dnsListenAddr)
		},
	}
	_, err = net.DefaultResolver.LookupIPAddr(ctx, "github.com")
	if err != nil {
		return fmt.Errorf("resolving github.com: %w", err)
	}
	return nil
}
