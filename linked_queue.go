package concurrent_queue

import (
	"context"
	"sync/atomic"
	"unsafe"
)

type LinkedQueue[T any] struct {
	head  unsafe.Pointer
	tail  unsafe.Pointer
	count uint64
}

func NewLinkedQueue[T any]() *LinkedQueue[T] {
	head := &node[T]{}
	ptr := unsafe.Pointer(head)
	return &LinkedQueue[T]{
		head: ptr,
		tail: ptr,
	}
}

func (q *LinkedQueue[T]) Enqueue(ctx context.Context, data T) error {
	newNode := &node[T]{val: data}
	newNodePtr := unsafe.Pointer(newNode)

	// 先改 tail
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// select tail; => tail = 4
		tailPtr := atomic.LoadPointer(&q.tail)
		// 为什么不能这样写？
		// tail = c.tail // 这种是非线程安全
		// Update Set tail = 3 WHERE tail = 4
		if atomic.CompareAndSwapPointer(&q.tail, tailPtr, newNodePtr) {
			// 在这一步，就要讲 tail.next 指向 c.tail
			// tail.next = c.tail
			tailNode := (*node[T])(tailPtr)
			// 你在这一步，c.tail 被人修改了
			atomic.StorePointer(&tailNode.next, newNodePtr)
			atomic.AddUint64(&q.count, 1)
			return nil
		}

		//// 先改 tail.next
		//newNode := &node[T]{val: data}
		//newPtr := unsafe.Pointer(newNode)
		//for {
		//	tailPtr := atomic.LoadPointer(&q.tail)
		//	tail := (*node[T])(tailPtr)
		//	tailNextPtr := atomic.LoadPointer(&tail.next)
		//	if tailNextPtr != nil {
		//		// 已经被人修改了，我们不需要修复，因为预期中修改的那个人会把 c.tail 指过去
		//		continue
		//	}
		//	if atomic.CompareAndSwapPointer(&tail.next, tailNextPtr, newPtr) {
		//		// 如果失败也不用担心，说明有人抢先一步了
		//		atomic.CompareAndSwapPointer(&q.tail, tailPtr, newPtr)
		//		return nil
		//	}
		//}
	}
}

func (q *LinkedQueue[T]) Dequeue(ctx context.Context) (T, error) {
	for {
		if ctx.Err() != nil {
			var t T
			return t, ctx.Err()
		}
		headPtr := atomic.LoadPointer(&q.head)
		headNode := (*node[T])(headPtr)
		tailPtr := atomic.LoadPointer(&q.tail)
		tailNode := (*node[T])(tailPtr)
		if headNode == tailNode {
			// 不需要做更多检测，在当下这一刻，我们就认为没有元素，即便这时候正好有人入队
			// 但是并不妨碍我们在它彻底入队完成——即所有的指针都调整好——之前，
			// 认为其实还是没有元素
			var t T
			return t, ErrEmptyQueue
		}
		headNextPtr := atomic.LoadPointer(&headNode.next)
		// 如果到这里为空了，CAS 操作不会成功。因为原本的数据，被人拿走了
		if atomic.CompareAndSwapPointer(&q.head, headPtr, headNextPtr) {
			headNextNode := (*node[T])(headNextPtr)
			// TODO TestLinkedQueue 测试这一步貌似出现问题
			return headNextNode.val, nil
		}
	}
}

func (q *LinkedQueue[T]) IsFull() bool {
	// TODO implement me
	panic("implement me")
}

func (q *LinkedQueue[T]) IsEmpty() bool {
	return atomic.LoadUint64(&q.count) == 0
}

func (q *LinkedQueue[T]) Len() uint64 {
	// 在你读的过程中，就被人改了
	return atomic.LoadUint64(&q.count)
}

type node[T any] struct {
	next unsafe.Pointer
	val  T
}
