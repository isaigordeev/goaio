package goaio

import "context"

type Resource[T any] struct {
	Data     []T
	Streamer Streamer[T]
}

func (r *Resource[T]) IsStreamable() bool {
	return r.Streamer != nil
}

func (r *Resource[T]) Stream(ctx context.Context) (<-chan T, <-chan error) {
	return r.Streamer.Streamer()
}

func NewStreamResource[T any](s Streamer[T]) *Resource[T] {
	return &Resource[T]{Streamer: s}
}
