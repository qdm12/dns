package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/qdm12/dns/v2/pkg/doh"
	cachemiddleware "github.com/qdm12/dns/v2/pkg/middlewares/cache"
	"github.com/qdm12/dns/v2/pkg/middlewares/cache/lru"
	"github.com/qdm12/dns/v2/pkg/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	logger := new(Logger)

	dohDialer, err := doh.New(doh.Settings{})
	if err != nil {
		log.Fatal(err)
	}

	cache, err := lru.New(lru.Settings{})
	if err != nil {
		log.Fatal(err)
	}

	cacheMiddleware, err := cachemiddleware.New(cachemiddleware.Settings{Cache: cache})
	if err != nil {
		log.Fatal(err)
	}

	server, err := server.New(server.Settings{
		Dialer:      dohDialer,
		Middlewares: []server.Middleware{cacheMiddleware},
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
