package concurrent_queue

import (
	"context"
	"golang.org/x/sync/semaphore"
	"sync"
)

// ConcurrentArrayBlockingQueue 有界并发阻塞队列
type ConcurrentArrayBlockingQueue[T any] struct {
	data  []T
	mutex *sync.Mutex

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

type ConcurrentArrayBlockingQueueV1[T any] struct {
	data  []T
	mutex *sync.Mutex

	maxSize int

	notEmptyCond *cond
	notFullCond  *cond
}

func NewConcurrentArrayBlockingQueueV1[T any](capacity int) *ConcurrentArrayBlockingQueueV1[T] {
	m := &sync.Mutex{}
	res := &ConcurrentArrayBlockingQueueV1[T]{
		data:    make([]T, 0, capacity),
		mutex:   m,
		maxSize: capacity,
		notEmptyCond: &cond{
			Cond: sync.NewCond(m),
		},
		notFullCond: &cond{
			Cond: sync.NewCond(m),
		},
	}
	return res
}

func (c *ConcurrentArrayBlockingQueueV1[T]) Enqueue(ctx context.Context, t T) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.mutex.Lock()
	if c.isFull() {
		err := c.notFullCond.WaitTimeout(ctx)
		if err != nil {
			return err
		}
	}
	c.data = append(c.data, t)
	// 没有人等 notEmpty 的信号，这一句就会阻塞住
	c.notEmptyCond.Signal()
	c.mutex.Unlock()
	return nil
}

func (c *ConcurrentArrayBlockingQueueV1[T]) Dequeue(ctx context.Context) (T, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	c.mutex.Lock()
	if c.isEmpty() {
		err := c.notEmptyCond.WaitTimeout(ctx)
		if err != nil {
			return nil, err
		}
	}
	// 没有人等 notFull 的信号，这一句就会阻塞住
	t := c.data[0]
	c.data = c.data[1:]
	c.notFullCond.Signal()
	c.mutex.Unlock()
	return t, nil
}

func (c *ConcurrentArrayBlockingQueueV1[T]) IsFull() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isFull()
}

func (c *ConcurrentArrayBlockingQueueV1[T]) isFull() bool {
	return len(c.data) == c.maxSize
}

func (c *ConcurrentArrayBlockingQueueV1[T]) IsEmpty() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isEmpty()
}

func (c *ConcurrentArrayBlockingQueueV1[T]) isEmpty() bool {
	return len(c.data) == 0
}

func (c *ConcurrentArrayBlockingQueueV1[T]) Len() uint64 {
	return uint64(len(c.data))
}

type cond struct {
	*sync.Cond
}

func (c *cond) WaitTimeout(ctx context.Context) error {
	ch := make(chan struct{})
	go func() {
		c.Wait()
		select {
		case ch <- struct{}{}:
		default:
			// 这里已经超时返回了
			c.Signal()
			c.L.Unlock()
		}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
		// 真的被唤醒了
		return nil
	}
}

type ConcurrentArrayBlockingQueueV2[T any] struct {
	data  []T
	mutex *sync.Mutex

	maxSize int

	notEmptyCond *cond
	notFullCond  *cond
}

func NewConcurrentArrayBlockingQueueV2[T any](capacity int) *ConcurrentArrayBlockingQueueV2[T] {
	res := &ConcurrentArrayBlockingQueueV2[T]{}
	return res
}
