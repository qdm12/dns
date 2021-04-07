package models

import "context"

type Server interface {
	Run(ctx context.Context, crashed chan<- error)
}
