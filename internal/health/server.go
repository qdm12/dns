package health

import (
	"context"

	"github.com/qdm12/goservices/httpserver"
)

type Infoer interface {
	Info(s string)
}

func NewServer(address string, logger Infoer, dnsListenAddr string) (
	server *httpserver.Server, err error,
) {
	ctx, cancel := context.WithCancel(context.Background()) //nolint:gosec
	handler := newHandler(ctx, dnsListenAddr)
	settings := httpserver.Settings{
		Name:          new("health"),
		Address:       &address,
		Handler:       handler,
		Logger:        logger,
		CancelHandler: cancel,
	}
	return httpserver.New(settings)
}
