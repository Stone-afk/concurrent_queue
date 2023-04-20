package concurrent_queue

import (
	"context"
	"sync"
	"testing"
)

// 切片实现
// BenchmarkConcurrentQueue-12       100000               306.7 ns/op           223 B/op          4 allocs/op

// BenchmarkConcurrentQueue-12       100000               280.8 ns/op           208 B/op          4 allocs/op
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
