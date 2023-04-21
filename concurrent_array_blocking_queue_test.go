package concurrent_queue

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestConcurrentBlockingQueue(t *testing.T) {
	// 只能确保没有死锁
	q := NewConcurrentArrayBlockingQueueV2[int](10000)
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
// BenchmarkConcurrentQueue-8       2783775               413.3 ns/op
func BenchmarkConcurrentQueue(b *testing.B) {
	var wg sync.WaitGroup
	q := NewConcurrentArrayBlockingQueueV2[int](100)
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
