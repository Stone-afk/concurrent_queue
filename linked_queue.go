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

		// 通过原子操作把队尾拿出来
		// select tail; => tail = 4
		tailPtr := atomic.LoadPointer(&q.tail)

		// 为什么不能这样写？
		// tail = c.tail // 这种是非线程安全
		// Update Set tail = 3 WHERE tail = 4

		// CAS 操作，如果当前的队尾指针就是上面取到的指针，那么把队尾换成新的结点
		if atomic.CompareAndSwapPointer(&q.tail, tailPtr, newNodePtr) {

			// CAS 返回成功，说明队尾没变，可以直接修改
			// 把新结点接到原来的队尾结点上去

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

		// CAS 返回失败，说明队尾变了，其他想要入队的，已经抢先入队而且完成了，那就要重头再来
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

		//   检查队列是否为空;
		if headNode == tailNode {
			// 不需要做更多检测，在当下这一刻，我们就认为没有元素，即便这时候正好有人入队
			// 但是并不妨碍我们在它彻底入队完成——即所有的指针都调整好——之前，
			// 认为其实还是没有元素
			var t T
			return t, ErrEmptyQueue
		}

		// 通过原子操作把队头节点拿出来
		headNextPtr := atomic.LoadPointer(&headNode.next)

		// CAS 操作 (如果当前的队头节点就是上面取到的节点，那么把队头换成当前队头节点的下一个节
		if atomic.CompareAndSwapPointer(&q.head, headPtr, headNextPtr) {

			//CAS 返回成功, 说明队头没变，可以直接修改, 把对头换成当前对头节点的下一个节点,
			// 返回当前对头节点。

			headNextNode := (*node[T])(headNextPtr)
			// TODO TestLinkedQueue 测试这一步貌似出现问题
			return headNextNode.val, nil
		}
	}

	// CAS 返回失败，说明队头变了，其他想要出队的，已经抢先出队而且完成了，那就要重头再来
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
