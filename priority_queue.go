package concurrent_queue

import "concurrent_queue/errs"

// RealNumber 实数
// 绝大多数情况下，你都应该用这个来表达数字的含义
type RealNumber interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64
}

type Number interface {
	RealNumber | ~complex64 | ~complex128
}

// Comparator 用于比较两个对象的大小 src < dst, 返回-1，src = dst, 返回0，src > dst, 返回1
type Comparator[T any] func(src T, dst T) int

func ComparatorRealNumber[T RealNumber](src T, dst T) int {
	if src < dst {
		return -1
	} else if src == dst {
		return 0
	} else {
		return 1
	}
}

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

func (p *PriorityQueue[T]) IsBoundless() bool {
	return p.capacity <= 0
}

func (p *PriorityQueue[T]) shrinkIfNecessary() {
	if p.IsBoundless() {
		p.data = Shrink[T](p.data)
	}
}

func (p *PriorityQueue[T]) Len() int {
	return len(p.data) - 1
}

func (p *PriorityQueue[T]) Cap() int {
	return p.capacity
}

func (p *PriorityQueue[T]) IsEmpty() bool {
	return len(p.data) < 2
}

func (p *PriorityQueue[T]) IsFull() bool {
	return p.capacity > 0 && len(p.data)-1 == p.capacity
}

func (p *PriorityQueue[T]) Peek() (T, error) {
	if p.IsEmpty() {
		var t T
		return t, errs.ErrEmptyQueue
	}
	return p.data[1], nil
}

func (p *PriorityQueue[T]) Enqueue(t T) error {
	if p.IsFull() {
		return errs.ErrOutOfCapacity
	}

	p.data = append(p.data, t)
	node, parent := len(p.data)-1, (len(p.data)-1)/2
	for parent > 0 && p.compare(p.data[node], p.data[parent]) < 0 {
		p.data[parent], p.data[node] = p.data[node], p.data[parent]
		node = parent
		parent = parent / 2
	}

	return nil
}

// heapify 下沉节点
func (p *PriorityQueue[T]) heapify(data []T, n, i int) {
	minPos := i
	for {
		if left := i * 2; left <= n && p.compare(data[left], data[minPos]) < 0 {
			minPos = left
		}
		if right := i*2 + 1; right <= n && p.compare(data[right], data[minPos]) < 0 {
			minPos = right
		}
		if minPos == i {
			break
		}
		data[i], data[minPos] = data[minPos], data[i]
		i = minPos
	}
}

func (p *PriorityQueue[T]) Dequeue() (T, error) {
	if p.IsEmpty() {
		var t T
		return t, errs.ErrEmptyQueue
	}
	pop := p.data[1]
	p.data[1] = p.data[len(p.data)-1]
	p.data = p.data[:len(p.data)-1]
	p.shrinkIfNecessary()
	p.heapify(p.data, len(p.data)-1, 1)
	return pop, nil
}

func Shrink[T any](src []T) []T {
	c, l := cap(src), len(src)
	n, changed := calCapacity(c, l)
	if !changed {
		return src
	}
	s := make([]T, 0, n)
	s = append(s, src...)
	return s
}

func calCapacity(c, l int) (int, bool) {
	if c <= 64 {
		return c, false
	}
	if c > 2048 && (c/l >= 2) {
		factor := 0.625
		return int(float32(c) * float32(factor)), true
	}
	if c <= 2048 && (c/l >= 4) {
		return c / 2, true
	}
	return c, false
}
