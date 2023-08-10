package dot

import (
	"context"
	"fmt"

	"github.com/qdm12/dns/v2/internal/server"
)

func newDNSHandler(ctx context.Context, settings ServerSettings) (
	handler *server.Handler, err error) {
	dial, err := newDoTDial(settings.Resolver)
	if err != nil {
		return nil, fmt.Errorf("creating DoT dial: %w", err)
	}

	exchange := server.NewExchange("DoT", dial, settings.Logger)

	return server.New(ctx, exchange, settings.Logger), nil
}
