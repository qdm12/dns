package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/qdm12/goservices"
)

type subServer struct {
	server *dns.Server

	startStopMutex sync.Mutex
	stateMutex     sync.RWMutex
	state          goservices.State
	done           <-chan struct{}
}

func newSubServer(server *dns.Server) *subServer {
	return &subServer{
		server: server,
	}
}

func (s *subServer) String() string {
	switch {
	case s.server.PacketConn != nil:
		return "DNS udp server"
	case s.server.Listener != nil:
		return "DNS tcp server"
	default:
		panic("unknown server type")
	}
}

func (s *subServer) Start(ctx context.Context) (runError <-chan error, err error) {
	s.startStopMutex.Lock()
	defer s.startStopMutex.Unlock()

	// Lock the state in case the subserver is already running.
	s.stateMutex.RLock()
	state := s.state
	// no need to keep a lock on the state since the `startStopMutex`
	// prevents concurrent calls to `Start` and `Stop`.
	s.stateMutex.RUnlock()
	if state == goservices.StateRunning {
		return nil, fmt.Errorf("%s: %w", s, goservices.ErrAlreadyStarted)
	}

	s.state = goservices.StateStarting

	ready := make(chan struct{})
	done := make(chan struct{})
	s.done = done
	runErrorCh := make(chan error)
	runError = runErrorCh

	go func(ready, done chan<- struct{},
		runError chan<- error,
	) {
		defer close(done)
		close(ready)
		err := s.server.ActivateAndServe()
		if err != nil {
			s.stateMutex.RLock()
			state := s.state
			s.stateMutex.RUnlock()
			if state == goservices.StateStopping {
				return
			}
			s.stateMutex.Lock()
			s.state = goservices.StateCrashed
			s.stateMutex.Unlock()
			runError <- err
		}
	}(ready, done, runErrorCh)

	<-ready

	const noCrashTimeout = 10 * time.Millisecond
	startTimer := time.NewTimer(noCrashTimeout)
	select {
	case <-ctx.Done(): // skip the timer
	case <-startTimer.C:
	case err = <-runError:
		return nil, err
	}

	s.stateMutex.Lock()
	s.state = goservices.StateRunning
	s.stateMutex.Unlock()

	return runError, nil
}

func (s *subServer) Stop() (err error) {
	s.startStopMutex.Lock()
	defer s.startStopMutex.Unlock()

	s.stateMutex.Lock()
	switch s.state {
	case goservices.StateRunning: // continue stopping the service
	case goservices.StateCrashed:
		s.stateMutex.Unlock()
		return fmt.Errorf("%w (crashed)", goservices.ErrAlreadyStopped)
	case goservices.StateStopped:
		s.stateMutex.Unlock()
		return fmt.Errorf("%w", goservices.ErrAlreadyStopped)
	case goservices.StateStarting, goservices.StateStopping:
		s.stateMutex.Unlock()
		panic("bad implementation code: " +
			"this code path should be unreachable for the \"" +
			fmt.Sprint(s.state) + "\" state")
	}
	s.state = goservices.StateStopping
	s.stateMutex.Unlock()

	err = s.server.Shutdown()
	s.state = goservices.StateStopped

	return err
}
