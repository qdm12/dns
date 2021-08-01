package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/qdm12/dns/pkg/cache"
	"github.com/qdm12/dns/pkg/doh"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	logger := new(Logger)
	server := doh.NewServer(ctx, logger, doh.ServerSettings{
		Cache: cache.Settings{Type: cache.LRU},
	})
	stopped := make(chan error)
	go server.Run(ctx, stopped)
	select {
	case <-ctx.Done():
		logger.Warn("\nCaught an OS signal, terminating...")
		<-stopped
	case err := <-stopped:
		logger.Warn("DoH server crashed: " + err.Error())
		stop() // stop custom handling of OS signals
		cancel()
	}
}

type Logger struct{}

func (l *Logger) Debug(s string) { log.Println(s) }
func (l *Logger) Info(s string)  { log.Println(s) }
func (l *Logger) Warn(s string)  { log.Println(s) }
func (l *Logger) Error(s string) { log.Println(s) }
