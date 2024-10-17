package doh

import (
	"context"

	"github.com/qdm12/dns/v2/internal/server"
)

func newDNSHandler(ctx context.Context, settings ServerSettings) (
	handler *server.Handler,
) {
	dial := newDoHDial(settings.Resolver)

	exchange := server.NewExchange("DoH", dial, settings.Logger)

	return server.New(ctx, exchange, settings.Logger)
}
