package concurrent_queue

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type DelayQueue[T Delayable] struct {
	pq            *PriorityQueue[T]
	mutex         *sync.Mutex
	dequeueSignal *cond
	enqueueSignal *cond
}

func NewDelayQueue[T Delayable](capacity int) *DelayQueue[T] {
	m := &sync.Mutex{}
	return &DelayQueue[T]{
		pq: NewPriorityQueue[T](capacity, func(src, dst T) int {
			srcDelay := src.Delay()
			dstDelay := dst.Delay()
			if srcDelay > dstDelay {
				return 1
			}
			if srcDelay == dstDelay {
				return 0
			}
			return -1
		}),
		mutex:         m,
		enqueueSignal: newCond(m),
		dequeueSignal: newCond(m),
	}
}

func (q *DelayQueue[T]) EnQueue(ctx context.Context, data T) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		// 跑过来这边，逻辑就是
		// 如果入队后的元素，过期时间更短，那么就要唤醒出队的
		// 或者，一点都不管，就直接唤醒出队的
		q.mutex.Lock()

		err := q.pq.Enqueue(data)
		switch err {
		case nil:
			// 入队成功
			// 发送入队信号，唤醒出队阻塞的

			// 优化
			// 如果新添加进来的元素，比原来堆顶元素deadline还早，说明是新的堆顶，则通知消费者堆顶变更了
			// if data.Deadline().Before(top.Deadline()) {
			//
			// }
			q.enqueueSignal.broadcast()
			return nil
		case ErrOutOfCapacity:
			signalCh := q.dequeueSignal.signalCh()
			// 阻塞，开始睡觉了
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-signalCh:
			}
		default:
			q.mutex.Unlock()
			return fmt.Errorf("延时队列入队的时候遇到未知错误 %w，请上报", err)
		}

	}
}

func (q *DelayQueue[T]) DeQueue(ctx context.Context) (T, error) {
	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()
	for {
		select {
		case <-ctx.Done():
			var t T
			return t, ctx.Err()
		default:
		}

		q.mutex.Lock()
		// 主要是顾虑锁被人持有很久，以至于早就超时了
		select {
		case <-ctx.Done():
			var t T
			q.mutex.Unlock()
			return t, ctx.Err()
		default:
		}
		val, err := q.pq.Peek()
		switch err {
		case nil:
			delayTime := val.Delay()
			if delayTime <= 0 {
				val, err = q.pq.Dequeue()
				if err != nil {
					var t T
					q.mutex.Unlock()
					return t, err
				}
				q.dequeueSignal.broadcast()
				return val, nil
			}
			// 要在这里解锁
			signalCh := q.enqueueSignal.signalCh()
			if timer == nil {
				timer = time.NewTimer(delayTime)
			} else {
				timer.Reset(delayTime)
			}
			select {
			case <-ctx.Done():
				var t T
				return t, ctx.Err()
			case <-timer.C:
				// 在这里不能这么写，因为这里已经无锁保护了
				// c.mu.Lock()
				// val, err = c.pq.Dequeue()
				// c.mu.Unlock()
				// return val, err
			case <-signalCh:
			}
		case ErrEmptyQueue:
			signalCh := q.enqueueSignal.signalCh()
			// 阻塞，开始 sleep
			select {
			case <-ctx.Done():
				var t T
				return t, ctx.Err()
			case <-signalCh:
			}
		default:
			q.mutex.Unlock()
			var t T
			return t, fmt.Errorf("延时队列出队的时候遇到未知错误 %w，请上报", err)
		}
	}
}

type cond struct {
	signal chan struct{}
	l      sync.Locker
}

func newCond(l sync.Locker) *cond {
	return &cond{
		l:      l,
		signal: make(chan struct{}),
	}
}

// broadcast 唤醒等待者
// 如果没有人等待，那么什么也不会发生
// 必须加锁之后才能调用这个方法
// 广播之后锁会被释放，这也是为了确保用户必然是在锁范围内调用的
func (c *cond) broadcast() {
	signal := make(chan struct{})
	ch := c.signal
	c.signal = signal
	c.l.Unlock()
	close(ch)
}

// signalCh 返回一个 channel，用于监听广播信号
// 必须在锁范围内使用
// 调用后，锁会被释放，这也是为了确保用户必然是在锁范围内调用的
func (c *cond) signalCh() <-chan struct{} {
	ch := c.signal
	c.l.Unlock()
	return ch
}
