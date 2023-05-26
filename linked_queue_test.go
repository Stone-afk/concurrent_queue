package concurrent_queue

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"testing"
)

func TestCAS(t *testing.T) {
	// a := 1
	// b := 2
	// a,b  = b,a => tmp = a, a=b, b=tmp
	// 是cas么?
	// 完全不是

	var value int64 = 10
	// 我准备把 value 更新为 12，当且仅当 value 原本的值是 10
	res := atomic.CompareAndSwapInt64(&value, 10, 12)

	// 这个不是并发安全的，要么就是利用锁，要么就是我们刚才的 CAS
	value = 12

	// res := atomic.CompareAndSwapInt64(&value, 11, 12)
	log.Println(res)
	log.Println(value)
}

func (q *LinkedQueue[T]) asSlice() []T {
	var res []T
	//curPointer := (*node[T])(q.head).next
	//cur := (*node[T])(curPointer)
	cur := (*node[T])((*node[T])(q.head).next)
	for cur != nil {
		res = append(res, cur.val)
		cur = (*node[T])(cur.next)
	}
	return res
}

func ExampleNewLinkedQueue() {
	q := NewLinkedQueue[int]()
	_ = q.Enqueue(context.Background(), 10)
	val, err := q.Dequeue(context.Background())
	if err != nil {
		// 一般意味着队列为空
		fmt.Println(err)
	}
	fmt.Println(val)
	// Output:
	// 10
}
