package health

import (
	"github.com/qdm12/dns/v2/internal/httpserver"
)

type Logger interface {
	Info(s string)
}

func NewServer(address string, healthcheck func() error) (
	server *httpserver.Server, err error) {
	handler := newHandler(healthcheck)
	settings := httpserver.Settings{
		Name:    stringPtr("health"),
		Address: &address,
		Handler: handler,
	}
	return httpserver.New(settings)
}

func stringPtr(s string) *string { return &s }
