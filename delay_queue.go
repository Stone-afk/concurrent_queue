package concurrent_queue

import (
	"context"
	"sync"
)

type DelayQueue[T Delayable] struct {
	pq            *PriorityQueue[T]
	mu            sync.RWMutex
	dequeueSignal *cond
	enqueueSignal *cond
}

func NewDelayQueue[T Delayable](capacity int) *DelayQueue[T] {
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
	}
}

func (q *DelayQueue[T]) EnQueue(ctx context.Context, data T) error {
	// TODO implement me
	panic("implement me")
}

func (q *DelayQueue[T]) DeQueue(ctx context.Context) error {
	// TODO implement me
	panic("implement me")
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

func (c *cond) broadcast() {
	// TODO implement me
	panic("implement me")
}

func (c *cond) signalCh() <-chan struct{} {
	// TODO implement me
	panic("implement me")
}
