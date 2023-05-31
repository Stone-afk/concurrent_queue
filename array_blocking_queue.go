package concurrent_queue

import (
	"context"
	"golang.org/x/sync/semaphore"
	"sync"
	"sync/atomic"
	"unsafe"
)

// ArrayBlockingQueue 有界并发阻塞队列
type ArrayBlockingQueue[T any] struct {
	data []T

	// 队头元素下标
	head int
	// 队尾元素下标
	tail int
	// 包含多少个元素
	count int
	// zero 不能作为返回值返回，防止用户篡改
	zero T

	mutex *sync.RWMutex

	enqueueCap *semaphore.Weighted
	dequeueCap *semaphore.Weighted
}

// NewArrayBlockingQueue 创建一个有界阻塞队列
// 容量会在最开始的时候就初始化好
// capacity 必须为正数
func NewArrayBlockingQueue[T any](capacity int) *ArrayBlockingQueue[T] {
	mutex := &sync.RWMutex{}

	semaForEnqueue := semaphore.NewWeighted(int64(capacity))
	semaForDequeue := semaphore.NewWeighted(int64(capacity))

	// error暂时不处理，因为目前没办法处理，只能考虑panic掉
	// 相当于将信号量置空
	_ = semaForDequeue.Acquire(context.TODO(), int64(capacity))

	res := &ArrayBlockingQueue[T]{
		data:       make([]T, capacity),
		mutex:      mutex,
		enqueueCap: semaForEnqueue,
		dequeueCap: semaForDequeue,
	}
	return res
}

func (q *ArrayBlockingQueue[T]) Enqueue(ctx context.Context, t T) error {

	// 能拿到，说明队列还有空位，可以入队，拿不到则阻塞
	err := q.enqueueCap.Acquire(ctx, 1)
	if err != nil {
		return err
	}

	q.mutex.Lock()
	// 拿到锁，先判断是否超时，防止在抢锁时已经超时
	if ctx.Err() != nil {
		// 超时应该主动归还信号量，避免容量泄露
		q.enqueueCap.Release(1)
		return ctx.Err()
	}

	q.data[q.tail] = t
	q.tail++
	q.count++

	// c.tail 已经是最后一个了，重置下标
	if q.tail == cap(q.data) {
		q.tail = 0
	}

	// 往出队的sema放入一个元素，出队的goroutine可以拿到并出队
	q.dequeueCap.Release(1)

	q.mutex.Unlock()

	return nil
}

func (q *ArrayBlockingQueue[T]) Dequeue(ctx context.Context) (T, error) {
	// 能拿到，说明队列有元素可以取，可以出队，拿不到则阻塞
	err := q.dequeueCap.Acquire(ctx, 1)

	var t T
	if err != nil {
		return t, err
	}

	q.mutex.Lock()

	// 拿到锁，先判断是否超时，防止在抢锁时已经超时
	if ctx.Err() != nil {
		// 超时应该主动归还信号量，有元素消费不到
		q.dequeueCap.Release(1)
		return t, ctx.Err()
	}

	t = q.data[q.head]
	// 为了释放内存，GC
	q.data[q.head] = q.zero
	q.head++
	q.count--
	if q.head == cap(q.data) {
		q.head = 0
	}

	// 往入队的sema放入一个元素，入队的goroutine可以拿到并入队
	q.enqueueCap.Release(1)
	q.mutex.Unlock()

	return t, nil
}

func (q *ArrayBlockingQueue[T]) IsFull() bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return q.isFull()
}

func (q *ArrayBlockingQueue[T]) isFull() bool {
	return q.count == cap(q.data)
}

func (q *ArrayBlockingQueue[T]) IsEmpty() bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return q.isEmpty()
}

func (q *ArrayBlockingQueue[T]) isEmpty() bool {
	return q.count == 0
}

func (q *ArrayBlockingQueue[T]) Len() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return q.count
}

type ArrayBlockingQueueV1[T any] struct {
	data  []T
	mutex *sync.Mutex

	maxSize int

	notEmptyCond *condV1
	notFullCond  *condV1
}

func NewArrayBlockingQueueV1[T any](capacity int) *ArrayBlockingQueueV1[T] {
	m := &sync.Mutex{}
	res := &ArrayBlockingQueueV1[T]{
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

func (c *ArrayBlockingQueueV1[T]) Enqueue(ctx context.Context, t T) error {
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

func (c *ArrayBlockingQueueV1[T]) Dequeue(ctx context.Context) (T, error) {
	if ctx.Err() != nil {
		var t T
		return t, ctx.Err()
	}
	c.mutex.Lock()
	if c.isEmpty() {
		err := c.notEmptyCond.WaitTimeout(ctx)
		if err != nil {
			var t T
			return t, err
		}
	}
	// 没有人等 notFull 的信号，这一句就会阻塞住
	t := c.data[0]
	c.data = c.data[1:]
	c.notFullCond.Signal()
	c.mutex.Unlock()
	return t, nil
}

func (c *ArrayBlockingQueueV1[T]) IsFull() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isFull()
}

func (c *ArrayBlockingQueueV1[T]) isFull() bool {
	return len(c.data) == c.maxSize
}

func (c *ArrayBlockingQueueV1[T]) IsEmpty() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isEmpty()
}

func (c *ArrayBlockingQueueV1[T]) isEmpty() bool {
	return len(c.data) == 0
}

func (c *ArrayBlockingQueueV1[T]) Len() uint64 {
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

type ArrayBlockingQueueV2[T any] struct {
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

func NewArrayBlockingQueueV2[T any](capacity int) *ArrayBlockingQueueV2[T] {
	m := &sync.RWMutex{}
	res := &ArrayBlockingQueueV2[T]{
		// 即便是 ring buffer，一次性分配完内存，也是有缺陷的
		// 如果不想一开始就把所有的内存都分配好，可以用链表
		data:         make([]T, capacity),
		mutex:        m,
		maxSize:      capacity,
		notEmptyCond: NewCond(m),
		notFullCond:  NewCond(m),
	}
	return res
}

func (c *ArrayBlockingQueueV2[T]) Enqueue(ctx context.Context, data T) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.mutex.Lock()
	for c.isFull() {
		err := c.notFullCond.WaitWithTimeout(ctx)
		if err != nil {
			return err
		}
	}

	c.data[c.tail] = data
	c.tail++
	c.count++
	if c.tail == c.maxSize {
		c.tail = 0
	}

	c.notEmptyCond.Broadcast()
	c.mutex.Unlock()
	return nil
}

func (c *ArrayBlockingQueueV2[T]) Dequeue(ctx context.Context) (T, error) {
	if ctx.Err() != nil {
		var t T
		return t, ctx.Err()
	}
	c.mutex.Lock()
	for c.isEmpty() {
		if err := c.notEmptyCond.WaitWithTimeout(ctx); err != nil {
			var t T
			return t, err
		}
	}

	t := c.data[c.head]
	c.data[c.head] = c.zero
	c.head++
	c.count--
	if c.head == c.maxSize {
		c.head = 0
	}
	c.notFullCond.Broadcast()
	c.mutex.Unlock()
	// 没有人等 notFull 的信号，这一句就会阻塞住
	return t, nil
}

func (c *ArrayBlockingQueueV2[T]) IsFull() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isFull()
}

func (c *ArrayBlockingQueueV2[T]) isFull() bool {
	return c.count == c.maxSize
}

func (c *ArrayBlockingQueueV2[T]) IsEmpty() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isEmpty()
}

func (c *ArrayBlockingQueueV2[T]) isEmpty() bool {
	return c.count == 0
}

func (c *ArrayBlockingQueueV2[T]) Len() uint64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return uint64(c.count)
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
