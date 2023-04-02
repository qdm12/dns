package services

import (
	"errors"
	"fmt"
)

var (
	ErrServiceIsNil = errors.New("service is nil")

	ErrNoService = errors.New("no service specified")

	ErrNoServiceStart            = errors.New("no service start order specified")
	ErrNoServiceStop             = errors.New("no service stop order specified")
	ErrServicesStartStopMismatch = errors.New("services to start and stop mismatch")
	ErrServicesNotUnique         = errors.New("services are not unique")

	ErrAlreadyStarted = errors.New("already started")
	ErrAlreadyStopped = errors.New("already stopped")
)

const (
	errorFormatCrash = "%s crashed: %s"
	errorFormatStart = "starting %s: %s"
	errorFormatStop  = "stopping %s: %s"
)

var _ error = serviceError{}

type serviceError struct {
	format      string
	serviceName string
	err         error
}

func (s serviceError) Error() string {
	if s.err == nil {
		panic("cannot have nil error in serviceError")
	}
	return fmt.Sprintf(s.format, s.serviceName, s.err.Error())
}

func (s serviceError) Unwrap() error {
	return s.err
}
