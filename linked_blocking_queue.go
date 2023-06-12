package concurrent_queue

import (
	"context"
	"github.com/ecodeclub/ekit/list"
	"sync"
)

type LinkedBlockingQueue[T any] struct {
	mutex *sync.RWMutex

	// 最大容量
	maxSize int
	// 链表
	linkedlist *list.LinkedList[T]

	notEmpty *cond
	notFull  *cond
}

func NewLinkedBlockingQueue[T any](capacity int) *LinkedBlockingQueue[T] {
	mutex := &sync.RWMutex{}
	return &LinkedBlockingQueue[T]{
		mutex:      mutex,
		maxSize:    capacity,
		linkedlist: list.NewLinkedList[T](),
	}
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
