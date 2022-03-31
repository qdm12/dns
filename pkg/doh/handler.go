package doh

import (
	"context"
	"fmt"

	"github.com/qdm12/dns/v2/internal/server"
)

func newDNSHandler(ctx context.Context, settings ServerSettings) (
	handler *server.Handler, err error) {
	dial, err := newDoHDial(settings.Resolver)
	if err != nil {
		return nil, fmt.Errorf("cannot create DoH dial: %w", err)
	}

	exchange := server.NewExchange("DoH", dial, settings.Logger)

	return server.New(ctx, exchange, settings.Filter,
		settings.Cache, settings.Logger), nil
}
