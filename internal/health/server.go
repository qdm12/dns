package health

import (
	"context"

	"github.com/qdm12/goservices/httpserver"
)

type Infoer interface {
	Info(s string)
}

func NewServer(address string, logger Infoer, healthcheck func(context.Context) error) (
	server *httpserver.Server, err error,
) {
	ctx, cancel := context.WithCancel(context.Background())
	handler := newHandler(ctx, healthcheck)
	settings := httpserver.Settings{
		Name:          stringPtr("health"),
		Address:       &address,
		Handler:       handler,
		Logger:        logger,
		CancelHandler: cancel,
	}
	return httpserver.New(settings)
}

func stringPtr(s string) *string { return &s }
