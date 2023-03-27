package concurrent_queue

import (
	"context"
	"golang.org/x/sync/semaphore"
	"sync"
)

// ConcurrentArrayBlockingQueue 有界并发阻塞队列
type ConcurrentArrayBlockingQueue[T any] struct {
	data  []T
	mutex *sync.RWMutex

	enqueueCap *semaphore.Weighted
	dequeueCap *semaphore.Weighted
}

// NewConcurrentArrayBlockingQueue 创建一个有界阻塞队列
// 容量会在最开始的时候就初始化好
// capacity 必须为正数
func NewConcurrentArrayBlockingQueue[T any](capacity int) *ConcurrentArrayBlockingQueue[T] {
	res := &ConcurrentArrayBlockingQueue[T]{}
	return res
}

func (c *ConcurrentArrayBlockingQueue[T]) Enqueue(ctx context.Context, t T) error {
	panic("")
}

func (c *ConcurrentArrayBlockingQueue[T]) Dequeue(ctx context.Context) (T, error) {
	panic("")
}
