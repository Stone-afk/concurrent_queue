package concurrent_queue

import (
	"context"
	"sync/atomic"
	"unsafe"
)

type ConcurrentLinkBlockingQueue[T any] struct {
	head  unsafe.Pointer
	tail  unsafe.Pointer
	count uint64
}

func NewConcurrentLinkedQueue[T any]() *ConcurrentLinkBlockingQueue[T] {
	head := &node[T]{}
	ptr := unsafe.Pointer(head)
	return &ConcurrentLinkBlockingQueue[T]{
		head: ptr,
		tail: ptr,
	}
}

func (c *ConcurrentLinkBlockingQueue[T]) Enqueue(ctx context.Context, data T) error {
	newNode := &node[T]{val: data}
	newNodePtr := unsafe.Pointer(newNode)

	// 先改 tail
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		tail := atomic.LoadPointer(&c.tail)
		if atomic.CompareAndSwapPointer(&c.tail, tail, newNodePtr) {
			tailNode := (*node[T])(tail)
			atomic.StorePointer(&tailNode.next, newNodePtr)
			atomic.AddUint64(&c.count, 1)
			return nil
		}
	}
}

func (c *ConcurrentLinkBlockingQueue[T]) Dequeue(ctx context.Context) (T, error) {
	// TODO implement me
	panic("implement me")
}

func (c *ConcurrentLinkBlockingQueue[T]) IsFull() bool {
	// TODO implement me
	panic("implement me")
}

func (c *ConcurrentLinkBlockingQueue[T]) IsEmpty() bool {
	// TODO implement me
	panic("implement me")
}

func (c *ConcurrentLinkBlockingQueue[T]) Len() uint64 {
	// 在你读的过程中，就被人改了
	return atomic.LoadUint64(&c.count)
}

type node[T any] struct {
	next unsafe.Pointer
	val  T
}
