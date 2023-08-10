package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/qdm12/dns/v2/pkg/cache/lru"
	cachemiddleware "github.com/qdm12/dns/v2/pkg/cache/middleware"
	"github.com/qdm12/dns/v2/pkg/dot"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	logger := new(Logger)

	cacheMiddleware := cachemiddleware.New(lru.New(lru.Settings{}))

	server, err := dot.NewServer(dot.ServerSettings{
		Middlewares: []dot.Middleware{cacheMiddleware},
		Logger:      logger,
	})
	if err != nil {
		log.Fatal(err)
	}

	runError, err := server.Start()
	if err != nil {
		log.Fatal(err)
	}

	select {
	case <-ctx.Done():
		logger.Warn("Caught an OS signal, terminating...")
		err = server.Stop()
		if err != nil {
			log.Fatal(err)
		}
		return
	case err := <-runError:
		logger.Warn("DoH server crashed: " + err.Error())
	}
}

type Logger struct{}

func (l *Logger) Debug(s string) { log.Println(s) }
func (l *Logger) Info(s string)  { log.Println(s) }
func (l *Logger) Warn(s string)  { log.Println(s) }
func (l *Logger) Error(s string) { log.Println(s) }
