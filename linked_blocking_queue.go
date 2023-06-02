package concurrent_queue

import "context"

type LinkedBlockingQueue[T any] struct{}

func NewLinkedBlockingQueue[T any](capacity int) *LinkedBlockingQueue[T] {
	// TODO implement me
	panic("implement me")
}

func (q *LinkedBlockingQueue[T]) Enqueue(ctx context.Context, data T) error {
	// TODO implement me
	panic("implement me")
}

func (q *LinkedBlockingQueue[T]) Len() int {
	// TODO implement me
	panic("implement me")
}
