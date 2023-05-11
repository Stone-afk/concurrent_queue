package concurrent_queue

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
