package concurrent_queue

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriorityQueue_Enqueue(t *testing.T) {
	testCases := []struct {
		name     string
		capacity int
		data     []int
		element  int
		wantErr  error
	}{
		{
			name:     "有界空队列",
			capacity: 10,
			data:     []int{},
			element:  10,
		},
		{
			name:     "有界满队列",
			capacity: 6,
			data:     []int{6, 5, 4, 3, 2, 1},
			element:  10,
			wantErr:  ErrOutOfCapacity,
		},
		{
			name:     "有界非空不满队列",
			capacity: 12,
			data:     []int{6, 5, 4, 3, 2, 1},
			element:  10,
		},
		{
			name:     "无界空队列",
			capacity: 0,
			data:     []int{},
			element:  10,
		},
		{
			name:     "无界非空队列",
			capacity: 0,
			data:     []int{6, 5, 4, 3, 2, 1},
			element:  10,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q := priorityQueueOf(tc.capacity, tc.data, compare())
			require.NotNil(t, q)
			err := q.Enqueue(tc.element)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.capacity, q.Cap())
		})

	}
}

func TestPriorityQueue_Peek(t *testing.T) {
	testCases := []struct {
		name     string
		capacity int
		data     []int
		wantErr  error
	}{
		{
			name:     "有数据",
			capacity: 0,
			data:     []int{6, 5, 4, 3, 2, 1},
			wantErr:  ErrEmptyQueue,
		},
		{
			name:     "无数据",
			capacity: 0,
			data:     []int{},
			wantErr:  ErrEmptyQueue,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q := NewPriorityQueue[int](tc.capacity, compare())
			for _, el := range tc.data {
				err := q.Enqueue(el)
				require.NoError(t, err)
			}
			for q.Len() > 0 {
				peek, err := q.Peek()
				assert.NoError(t, err)
				el, _ := q.Dequeue()
				assert.Equal(t, el, peek)
			}
			_, err := q.Peek()
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestNewPriorityQueue(t *testing.T) {
	data := []int{6, 5, 4, 3, 2, 1}
	testCases := []struct {
		name     string
		q        *PriorityQueue[int]
		capacity int
		data     []int
		expected []int
	}{
		{
			name:     "无边界",
			q:        NewPriorityQueue(0, compare()),
			capacity: 0,
			data:     data,
			expected: []int{1, 2, 3, 4, 5, 6},
		},
		{
			name:     "有边界 ",
			q:        NewPriorityQueue(len(data), compare()),
			capacity: len(data),
			data:     data,
			expected: []int{1, 2, 3, 4, 5, 6},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, 0, tc.q.Len())
			for _, d := range data {
				err := tc.q.Enqueue(d)
				assert.NoError(t, err)
				if err != nil {
					return
				}
			}
			assert.Equal(t, tc.capacity, tc.q.Cap())
			assert.Equal(t, len(data), tc.q.Len())
			res := make([]int, 0, len(data))
			for tc.q.Len() > 0 {
				el, err := tc.q.Dequeue()
				assert.NoError(t, err)
				if err != nil {
					return
				}
				res = append(res, el)
			}
			assert.Equal(t, tc.expected, res)
		})
	}
}

func compare() Comparator[int] {
	return func(a, b int) int {
		if a < b {
			return -1
		}
		if a == b {
			return 0
		}
		return 1
	}
}

func priorityQueueOf(capacity int, data []int, compare Comparator[int]) *PriorityQueue[int] {
	q := NewPriorityQueue[int](capacity, compare)
	for _, el := range data {
		err := q.Enqueue(el)
		if err != nil {
			return nil
		}
	}
	return q
}
