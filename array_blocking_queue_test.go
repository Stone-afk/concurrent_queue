package concurrent_queue

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestArrayBlockingQueueV2(t *testing.T) {
	// 只能确保没有死锁
	q := NewArrayBlockingQueueV2[int](10000)
	// data := make(chan int, 10000000000000000000000)
	// 并发的问题都落在 m 上
	// var m sync.Mutex
	var wg sync.WaitGroup
	wg.Add(30)
	for i := 0; i < 20; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				// 你没有办法校验这里面的中间结果
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				val := rand.Int()
				_ = q.Enqueue(ctx, val)
				cancel()
			}
			wg.Done()
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				_, _ = q.Dequeue(ctx)
				cancel()
			}
			wg.Done()
		}()
	}

	// 怎么校验 q 对还是不对
	wg.Wait()
}

// 切片实现
// BenchmarkArrayBlockingQueueV2-8       2783775               413.3 ns/op
func BenchmarkArrayBlockingQueueV2(b *testing.B) {
	var wg sync.WaitGroup
	q := NewArrayBlockingQueueV2[int](100)
	wg.Add(2)
	b.ResetTimer()
	go func() {
		for i := 0; i < b.N; i++ {
			_ = q.Enqueue(context.Background(), i)
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < b.N; i++ {
			_, _ = q.Dequeue(context.Background())
		}
		wg.Done()
	}()
	wg.Wait()
}

func ExampleNewArrayBlockingQueue() {
	q := NewArrayBlockingQueue[int](10)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = q.Enqueue(ctx, 22)
	val, err := q.Dequeue(ctx)
	// 这是例子，实际中你不需要写得那么复杂
	switch err {
	case context.Canceled:
		// 有人主动取消了，即调用了 cancel 方法。在这个例子里不会出现这个情况
	case context.DeadlineExceeded:
		// 超时了
	case nil:
		fmt.Println(val)
	default:
		// 其它乱七八糟的
	}
	// Output:
	// 22
}
