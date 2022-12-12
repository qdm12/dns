package health

import (
	"context"
	"net/http"
	"time"
)

var _ Runner = (*Server)(nil)

type Runner interface {
	Run(ctx context.Context, done chan<- struct{})
}

type Logger interface {
	Info(s string)
	Warn(s string)
	Error(s string)
}

type Server struct {
	address string
	logger  Logger
	handler http.Handler
}

func NewServer(address string, logger Logger, healthcheck func() error) *Server {
	handler := newHandler(healthcheck)
	return &Server{
		address: address,
		logger:  logger,
		handler: handler,
	}
}

func (s *Server) Run(ctx context.Context, done chan<- struct{}) {
	defer close(done)
	server := http.Server{Addr: s.address, Handler: s.handler}
	go func() {
		<-ctx.Done()
		s.logger.Warn("shutting down (context canceled)")
		defer s.logger.Warn("shut down")
		const shutdownGraceDuration = 2 * time.Second
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGraceDuration)
		defer cancel()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			s.logger.Error("failed shutting down: " + err.Error())
		}
	}()
	for ctx.Err() == nil {
		s.logger.Info("listening on " + s.address)
		err := server.ListenAndServe()
		if err != nil && ctx.Err() == nil { // server crashed
			s.logger.Error(err.Error())
			s.logger.Info("restarting")
		}
	}
}
