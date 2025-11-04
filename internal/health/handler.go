package health

import (
	"context"
	"net/http"
)

func newHandler(ctx context.Context, healthcheck func(context.Context) error) http.Handler {
	return &handler{
		ctx:         ctx,
		healthcheck: healthcheck,
	}
}

type handler struct {
	ctx         context.Context //nolint:containedctx
	healthcheck func(context.Context) error
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || (r.RequestURI != "" && r.RequestURI != "/") {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	err := h.healthcheck(h.ctx) //nolint:contextcheck
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
