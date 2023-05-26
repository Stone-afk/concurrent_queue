package concurrent_queue

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
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

func TestLinkedQueue(t *testing.T) {
	t.Parallel()
	// 仅仅是为了测试在入队出队期间不会出现 panic 或者死循环之类的问题
	// FIFO 特性参考其余测试
	q := NewLinkedQueue[int]()
	var wg sync.WaitGroup
	wg.Add(10000)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				val := rand.Int()
				_ = q.Enqueue(context.Background(), val)
			}
		}()
	}
	var cnt int32 = 0
	for i := 0; i < 10; i++ {
		go func() {
			for {
				if atomic.LoadInt32(&cnt) >= 10000 {
					return
				}
				_, err := q.Dequeue(context.Background())
				if err == nil {
					atomic.AddInt32(&cnt, 1)
					wg.Done()
				}
			}
		}()
	}
	wg.Wait()
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
