package prometheus

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	address string
	logger  Logger
	handler http.Handler
}

func (s *Server) Run(ctx context.Context, done chan<- struct{}) {
	defer close(done)

	waitCtx, waitCancel := context.WithCancel(context.Background())
	defer waitCancel()

	server := http.Server{
		Addr:    s.address,
		Handler: s.handler,
	}

	go func() {
		select {
		case <-ctx.Done():
		case <-waitCtx.Done():
			return
		}

		const shutdownGraceDuration = 2 * time.Second
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGraceDuration)
		defer cancel()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			s.logger.Error("failed shutting down: " + err.Error())
		}
	}()

	s.logger.Info("listening on " + s.address)
	err := server.ListenAndServe()
	if err != nil && ctx.Err() == nil {
		s.logger.Error(err.Error())
	}
}
