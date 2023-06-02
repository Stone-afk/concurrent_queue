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

func (q *LinkedBlockingQueue[T]) Dequeue(ctx context.Context) (T, error) {
	// TODO implement me
	panic("implement me")
}

func (q *LinkedBlockingQueue[T]) Len() int {
	// TODO implement me
	panic("implement me")
}

func (q *LinkedBlockingQueue[T]) IsEmpty() bool {
	// TODO implement me
	panic("implement me")
}

func (q *LinkedBlockingQueue[T]) IsFull() bool {
	// TODO implement me
	panic("implement me")
}
