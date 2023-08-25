// Package noop initializes all No-op metrics objects.
package noop

import "context"

type DummyRunner struct{}

func (d *DummyRunner) String() string {
	return "dummy metrics server"
}

func (d *DummyRunner) Start(context.Context) (runError <-chan error, startErr error) {
	return nil, nil //nolint:nilnil
}

func (d *DummyRunner) Stop() (err error) {
	return nil
}

func New() (dummy *DummyRunner, err error) {
	return &DummyRunner{}, nil
}
