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
	for {
		if ctx.Err() != nil {
			var t T
			return t, ctx.Err()
		}
		head := atomic.LoadPointer(&c.head)
		headNode := (*node[T])(head)
		tail := atomic.LoadPointer(&c.tail)
		tailNode := (*node[T])(tail)
		if headNode == tailNode {
			// 不需要做更多检测，在当下这一刻，我们就认为没有元素，即便这时候正好有人入队
			// 但是并不妨碍我们在它彻底入队完成——即所有的指针都调整好——之前，
			// 认为其实还是没有元素
			var t T
			return t, ErrEmptyQueue
		}
		headNext := atomic.LoadPointer(&headNode.next)
		// 如果到这里为空了，CAS 操作不会成功。因为原本的数据，被人拿走了
		if atomic.CompareAndSwapPointer(&c.head, head, headNext) {
			headNextNode := (*node[T])(headNext)
			return headNextNode.val, nil
		}
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
