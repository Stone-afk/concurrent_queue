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
	if ctx.Err() != nil {
		return ctx.Err()
	}
	q.mutex.Lock()
	for q.maxSize > 0 && q.isFull() {
		signal := q.notFull.signalCh()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-signal:
			q.mutex.Lock()
		}
	}
	err := q.linkedlist.Append(data)

	// 这里会释放锁
	q.notEmpty.broadcast()
	return err
}

// Dequeue 出队
// 注意：目前我们已经通过broadcast实现了超时控制
func (q *LinkedBlockingQueue[T]) Dequeue(ctx context.Context) (T, error) {
	if ctx.Err() != nil {
		var val T
		return val, ctx.Err()
	}
	q.mutex.Lock()
	for q.isEmpty() {
		signal := q.notEmpty.signalCh()
		select {
		case <-ctx.Done():
			var val T
			return val, ctx.Err()
		case <-signal:
			q.mutex.Lock()
		}
	}
	val, err := q.linkedlist.Delete(0)

	return val, err
}

func (q *LinkedBlockingQueue[T]) Len() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return q.len()
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

func (q *LinkedBlockingQueue[T]) AsSlice() []T {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return q.linkedlist.AsSlice()
}
