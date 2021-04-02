package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/qdm12/dns/pkg/dot"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	logger := new(Logger)
	server := dot.NewServer(ctx, logger, dot.ServerSettings{})
	stopped := make(chan struct{})
	go server.Run(ctx, stopped)
	select {
	case <-ctx.Done():
		logger.Warn("\nCaught an OS signal, terminating...")
	case <-stopped:
		logger.Warn("DoH server crashed")
		stop() // stop custom handling of OS signals
		cancel()
	}
	<-stopped
}

type Logger struct{}

func (l *Logger) Debug(args ...interface{}) { log.Println(args...) }
func (l *Logger) Info(args ...interface{})  { log.Println(args...) }
func (l *Logger) Warn(args ...interface{})  { log.Println(args...) }
func (l *Logger) Error(args ...interface{}) { log.Println(args...) }
