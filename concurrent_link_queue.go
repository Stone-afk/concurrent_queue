package concurrent_queue

import (
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
