package concurrent_queue

import "context"

// Comparator 用于比较两个对象的大小 src < dst, 返回-1，src = dst, 返回0，src > dst, 返回1
type Comparator[T any] func(src T, dst T) int

// PriorityQueue 是一个基于小顶堆的优先队列
// 当capacity= 0时，为无界队列，切片容量会动态扩缩容
// 当capacity!=0 时，为有界队列，初始化后就固定容量，不会扩缩容
type PriorityQueue[T any] struct {
	// 用于比较前一个元素是否小于后一个元素
	compare Comparator[T]
	// 队列容量
	capacity int
	// 队列中的元素，为便于计算父子节点的index，0位置留空，根节点从1开始
	data []T
}

// NewPriorityQueue 创建优先队列 capacity <= 0 时，为无界队列
func NewPriorityQueue[T any](capacity int, compare Comparator[T]) *PriorityQueue[T] {
	sliceCap := capacity + 1
	if capacity < 1 {
		capacity = 0
		sliceCap = 64
	}
	return &PriorityQueue[T]{
		capacity: capacity,
		compare:  compare,
		data:     make([]T, 1, sliceCap),
	}
}

func (p *PriorityQueue[T]) Len() int {
	// TODO implement me
	panic("implement me")
}

func (p *PriorityQueue[T]) Cap() int {
	// TODO implement me
	panic("implement me")
}

func (p *PriorityQueue[T]) IsEmpty() bool {
	// TODO implement me
	panic("implement me")
}

func (p *PriorityQueue[T]) IsFull() bool {
	// TODO implement me
	panic("implement me")
}

func (p *PriorityQueue[T]) Peek() (T, error) {
	// TODO implement me
	panic("implement me")
}

func (p *PriorityQueue[T]) Enqueue(ctx context.Context, t T) error {
	// TODO implement me
	panic("implement me")
}

func (p *PriorityQueue[T]) Dequeue(ctx context.Context) (T, error) {
	// TODO implement me
	panic("implement me")
}
