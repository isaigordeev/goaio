package goaio

import "context"

type Streamer[T any] interface {
	Streamer(ctx context.Context)	(<-chan T, <-error)
}


