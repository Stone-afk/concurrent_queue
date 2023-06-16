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
		notEmpty:   newCond(mutex),
		notFull:    newCond(mutex),
		linkedlist: list.NewLinkedList[T](),
	}
}

// Enqueue 入队
// 注意：目前我们已经通过broadcast实现了超时控制
func (q *LinkedBlockingQueue[T]) Enqueue(ctx context.Context, data T) error {
	// TODO implement me
	panic("implement me")
}

// Dequeue 出队
// 注意：目前我们已经通过broadcast实现了超时控制
func (q *LinkedBlockingQueue[T]) Dequeue(ctx context.Context) (T, error) {
	// TODO implement me
	panic("implement me")
}

func (q *LinkedBlockingQueue[T]) Len() int {
	// TODO implement me
	panic("implement me")
}

func (q *LinkedBlockingQueue[T]) len() int {
	return q.linkedlist.Len()
}

func (q *LinkedBlockingQueue[T]) IsEmpty() bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return q.isEmpty()
}

func (q *LinkedBlockingQueue[T]) isEmpty() bool {
	return q.linkedlist.Len() == 0
}

func (q *LinkedBlockingQueue[T]) IsFull() bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return q.isFull()
}

func (q *LinkedBlockingQueue[T]) isFull() bool {
	return q.linkedlist.Len() == q.maxSize
}
