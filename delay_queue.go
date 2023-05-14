package concurrent_queue

import "sync"

type DelayQueue[T Delayable] struct {
	pq            *PriorityQueue[T]
	mu            sync.RWMutex
	dequeueSignal *cond
	enqueueSignal *cond
}

type cond struct {
	signal chan struct{}
	l      sync.Locker
}
