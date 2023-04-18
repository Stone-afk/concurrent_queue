package concurrent_queue

import (
	"context"
	"golang.org/x/sync/semaphore"
	"sync"
	"sync/atomic"
	"unsafe"
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

	notEmptyCond *condV1
	notFullCond  *condV1
}

func NewConcurrentArrayBlockingQueueV1[T any](capacity int) *ConcurrentArrayBlockingQueueV1[T] {
	m := &sync.Mutex{}
	res := &ConcurrentArrayBlockingQueueV1[T]{
		data:    make([]T, 0, capacity),
		mutex:   m,
		maxSize: capacity,
		notEmptyCond: &condV1{
			Cond: sync.NewCond(m),
		},
		notFullCond: &condV1{
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

type condV1 struct {
	*sync.Cond
}

func (c *condV1) WaitTimeout(ctx context.Context) error {
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
	mutex *sync.RWMutex

	maxSize int

	notEmptyCond *CondV2
	notFullCond  *CondV2

	count int
	head  int
	tail  int

	zero T
}

func NewConcurrentArrayBlockingQueueV2[T any](capacity int) *ConcurrentArrayBlockingQueueV2[T] {
	m := &sync.RWMutex{}
	res := &ConcurrentArrayBlockingQueueV2[T]{
		data:         make([]T, 0, capacity),
		mutex:        m,
		maxSize:      capacity,
		notEmptyCond: NewCond(m),
		notFullCond:  NewCond(m),
	}
	return res
}

func (c *ConcurrentArrayBlockingQueueV2[T]) Enqueue(t T) error {
	panic("`Enqueue` must implemented")
}

func (c *ConcurrentArrayBlockingQueueV2[T]) Dequeue() (T, error) {
	panic("`Dequeue` must implemented")
}

func (c *ConcurrentArrayBlockingQueueV2[T]) IsFull() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isFull()
}

func (c *ConcurrentArrayBlockingQueueV2[T]) isFull() bool {
	return len(c.data) == c.maxSize
}

func (c *ConcurrentArrayBlockingQueueV2[T]) IsEmpty() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isEmpty()
}

func (c *ConcurrentArrayBlockingQueueV2[T]) isEmpty() bool {
	return len(c.data) == 0
}

func (c *ConcurrentArrayBlockingQueueV2[T]) Len() uint64 {
	return uint64(len(c.data))
}

type CondV2 struct {
	L  sync.Locker
	ch unsafe.Pointer
}

func NewCond(l sync.Locker) *CondV2 {
	c := &CondV2{L: l}
	ch := make(chan struct{})
	c.ch = unsafe.Pointer(&ch)
	return c
}

// Wait for Broadcast calls. Similar to regular sync.Cond, this unlocks the underlying
// locker first, waits on changes and re-locks it before returning.
func (c *CondV2) Wait() {
	ch := c.NotifyChan()
	c.L.Unlock()
	<-ch
	c.L.Lock()
}

// WaitWithTimeout Same as Wait() call, but will only wait up to a given timeout.
func (c *CondV2) WaitWithTimeout(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	ch := c.NotifyChan()
	c.L.Unlock()
	select {
	case <-ch:
		c.L.Lock()
		return nil
	case <-ctx.Done():
		c.L.Lock()
		return ctx.Err()
	}
}

// Broadcast call notifies everyone that something has changed.
func (c *CondV2) Broadcast() {
	ch := make(chan struct{})
	ptrOld := atomic.SwapPointer(&c.ch, unsafe.Pointer(&ch))
	// n := *((*chan struct{})(ptrOld))
	// close(n)
	close(*((*chan struct{})(ptrOld)))

}

// NotifyChan Returns a channel that can be used to wait for next Broadcast() call.
func (c *CondV2) NotifyChan() <-chan struct{} {
	uPtr := atomic.LoadPointer(&c.ch)
	//chPtr := (*chan struct{})(uPtr)
	//ch := *chPtr
	//return ch
	return *((*chan struct{})(uPtr))

}
