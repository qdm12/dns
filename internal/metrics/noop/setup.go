// Package noop initializes all No-op metrics objects.
package noop

import (
	"context"
)

type DummyRunner struct{}

func (d *DummyRunner) Run(ctx context.Context, done chan<- struct{}) {
	close(done)
}

func Setup() (dummy *DummyRunner) {
	return &DummyRunner{}
}
