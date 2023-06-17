package concurrent_queue

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestArrayBlockingQueueV2_Enqueue(t *testing.T) {
	testCases := []struct {
		name string

		q *ArrayBlockingQueueV2[int]

		timeout time.Duration
		value   int

		data []int

		wantErr error
	}{
		{
			name:    "enqueue",
			q:       NewArrayBlockingQueueV2[int](10),
			value:   1,
			timeout: time.Minute,
			data:    []int{1},
		},
		{
			name: "blocking and timeout",
			q: func() *ArrayBlockingQueueV2[int] {
				res := NewArrayBlockingQueueV2[int](2)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := res.Enqueue(ctx, 1)
				require.NoError(t, err)
				err = res.Enqueue(ctx, 2)
				require.NoError(t, err)
				return res
			}(),
			value:   3,
			timeout: time.Second,
			data:    []int{1, 2},
			wantErr: context.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()
			err := tc.q.Enqueue(ctx, tc.value)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.data, tc.q.data)
		})
	}
}

func TestArrayBlockingQueueV2(t *testing.T) {
	// 只能确保没有死锁
	q := NewArrayBlockingQueueV2[int](10000)
	// data := make(chan int, 10000000000000000000000)
	// 并发的问题都落在 m 上
	// var m sync.Mutex
	var wg sync.WaitGroup
	wg.Add(30)
	for i := 0; i < 20; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				// 你没有办法校验这里面的中间结果
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				val := rand.Int()
				_ = q.Enqueue(ctx, val)
				cancel()
			}
			wg.Done()
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				_, _ = q.Dequeue(ctx)
				cancel()
			}
			wg.Done()
		}()
	}

	// 怎么校验 q 对还是不对
	wg.Wait()
}

// 切片实现
// BenchmarkArrayBlockingQueueV2-8       2783775               413.3 ns/op
func BenchmarkArrayBlockingQueueV2(b *testing.B) {
	var wg sync.WaitGroup
	q := NewArrayBlockingQueueV2[int](100)
	wg.Add(2)
	b.ResetTimer()
	go func() {
		for i := 0; i < b.N; i++ {
			_ = q.Enqueue(context.Background(), i)
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < b.N; i++ {
			_, _ = q.Dequeue(context.Background())
		}
		wg.Done()
	}()
	wg.Wait()
}

func TestBlockingQueue_Enqueue(t *testing.T) {
	testCases := []struct {
		name      string
		q         func() *ArrayBlockingQueue[int]
		val       int
		timeout   time.Duration
		wantErr   error
		wantData  []int
		wantSlice []int
		wantLen   int
		wantHead  int
		wantTail  int
	}{
		{
			name: "empty and enqueued",
			q: func() *ArrayBlockingQueue[int] {
				return NewArrayBlockingQueue[int](3)
			},
			val:       123,
			timeout:   time.Second,
			wantData:  []int{123, 0, 0},
			wantSlice: []int{123},
			wantLen:   1,
			wantTail:  1,
			wantHead:  0,
		},
		{
			name: "invalid context",
			q: func() *ArrayBlockingQueue[int] {
				return NewArrayBlockingQueue[int](3)
			},
			val:       123,
			timeout:   -time.Second,
			wantData:  []int{0, 0, 0},
			wantSlice: []int{},
			wantErr:   context.DeadlineExceeded,
		},
		{
			// 入队之后就满了，恰好放在切片的最后一个位置
			name: "enqueued full last index",
			q: func() *ArrayBlockingQueue[int] {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				q := NewArrayBlockingQueue[int](3)
				err := q.Enqueue(ctx, 123)
				require.NoError(t, err)
				err = q.Enqueue(ctx, 234)
				require.NoError(t, err)
				return q
			},
			val:       345,
			timeout:   time.Second,
			wantData:  []int{123, 234, 345},
			wantSlice: []int{123, 234, 345},
			wantLen:   3,
			wantTail:  0,
			wantHead:  0,
		},
		{
			// 入队之后就满了，恰好放在切片的第一个
			name: "enqueued full middle index",
			q: func() *ArrayBlockingQueue[int] {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				q := NewArrayBlockingQueue[int](3)
				err := q.Enqueue(ctx, 123)
				require.NoError(t, err)
				err = q.Enqueue(ctx, 234)
				require.NoError(t, err)
				err = q.Enqueue(ctx, 345)
				require.NoError(t, err)
				val, err := q.Dequeue(ctx)
				require.NoError(t, err)
				require.Equal(t, 123, val)
				return q
			},
			val:       456,
			timeout:   time.Second,
			wantData:  []int{456, 234, 345},
			wantSlice: []int{234, 345, 456},
			wantLen:   3,
			wantTail:  1,
			wantHead:  1,
		},
		{
			// 入队之后就满了，恰好放在中间
			name: "enqueued full first index",
			q: func() *ArrayBlockingQueue[int] {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				q := NewArrayBlockingQueue[int](3)
				err := q.Enqueue(ctx, 123)
				require.NoError(t, err)
				err = q.Enqueue(ctx, 234)
				require.NoError(t, err)
				err = q.Enqueue(ctx, 345)
				require.NoError(t, err)
				val, err := q.Dequeue(ctx)
				require.NoError(t, err)
				require.Equal(t, 123, val)
				val, err = q.Dequeue(ctx)
				require.NoError(t, err)
				require.Equal(t, 234, val)
				err = q.Enqueue(ctx, 456)
				require.NoError(t, err)
				return q
			},
			val:       567,
			timeout:   time.Second,
			wantData:  []int{456, 567, 345},
			wantSlice: []int{345, 456, 567},
			wantLen:   3,
			wantTail:  2,
			wantHead:  2,
		},
		{
			// 元素本身就是零值
			name: "all zero value ",
			q: func() *ArrayBlockingQueue[int] {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				q := NewArrayBlockingQueue[int](3)
				err := q.Enqueue(ctx, 0)
				require.NoError(t, err)
				err = q.Enqueue(ctx, 0)
				require.NoError(t, err)
				return q
			},
			val:       0,
			timeout:   time.Second,
			wantData:  []int{0, 0, 0},
			wantSlice: []int{0, 0, 0},
			wantLen:   3,
			wantTail:  0,
			wantHead:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()
			q := tc.q()
			err := q.Enqueue(ctx, tc.val)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantData, q.data)
			assert.Equal(t, tc.wantSlice, q.AsSlice())
			assert.Equal(t, tc.wantLen, q.Len())
			assert.Equal(t, tc.wantHead, q.head)
			assert.Equal(t, tc.wantTail, q.tail)
		})
	}

	t.Run("enqueue timeout", func(t *testing.T) {
		q := NewArrayBlockingQueue[int](3)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := q.Enqueue(ctx, 123)
		require.NoError(t, err)
		err = q.Enqueue(ctx, 234)
		require.NoError(t, err)
		err = q.Enqueue(ctx, 345)
		require.NoError(t, err)
		err = q.Enqueue(ctx, 456)
		require.Equal(t, context.DeadlineExceeded, err)
	})

	// 入队阻塞，而后出队，于是入队成功
	t.Run("enqueue blocking and dequeue", func(t *testing.T) {
		q := NewArrayBlockingQueue[int](3)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		go func() {
			time.Sleep(time.Millisecond * 100)
			val, err := q.Dequeue(ctx)
			require.NoError(t, err)
			require.Equal(t, 123, val)
		}()
		err := q.Enqueue(ctx, 123)
		require.NoError(t, err)
		err = q.Enqueue(ctx, 234)
		require.NoError(t, err)
		err = q.Enqueue(ctx, 345)
		require.NoError(t, err)
		err = q.Enqueue(ctx, 456)
		require.NoError(t, err)
	})
}

func TestBlockingQueue_Dequeue(t *testing.T) {
	testCases := []struct {
		name      string
		q         func() *ArrayBlockingQueue[int]
		val       int
		timeout   time.Duration
		wantErr   error
		wantVal   int
		wantData  []int
		wantSlice []int
		wantLen   int
		wantHead  int
		wantTail  int
	}{
		{
			name: "dequeued",
			q: func() *ArrayBlockingQueue[int] {
				q := NewArrayBlockingQueue[int](3)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := q.Enqueue(ctx, 123)
				require.NoError(t, err)
				err = q.Enqueue(ctx, 234)
				require.NoError(t, err)
				return q
			},
			wantVal:   123,
			timeout:   time.Second,
			wantData:  []int{0, 234, 0},
			wantSlice: []int{234},
			wantLen:   1,
			wantHead:  1,
			wantTail:  2,
		},
		{
			name: "invalid context",
			q: func() *ArrayBlockingQueue[int] {
				q := NewArrayBlockingQueue[int](3)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := q.Enqueue(ctx, 123)
				require.NoError(t, err)
				err = q.Enqueue(ctx, 234)
				require.NoError(t, err)
				return q
			},
			wantErr:   context.DeadlineExceeded,
			timeout:   -time.Second,
			wantData:  []int{123, 234, 0},
			wantSlice: []int{123, 234},
			wantLen:   2,
			wantHead:  0,
			wantTail:  2,
		},
		{
			name: "dequeue and empty first",
			q: func() *ArrayBlockingQueue[int] {
				q := NewArrayBlockingQueue[int](3)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := q.Enqueue(ctx, 123)
				require.NoError(t, err)
				return q
			},
			wantVal:   123,
			timeout:   time.Second,
			wantData:  []int{0, 0, 0},
			wantSlice: []int{},
			wantLen:   0,
			wantHead:  1,
			wantTail:  1,
		},
		{
			name: "dequeue and empty middle",
			q: func() *ArrayBlockingQueue[int] {
				q := NewArrayBlockingQueue[int](3)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := q.Enqueue(ctx, 123)
				require.NoError(t, err)
				err = q.Enqueue(ctx, 234)
				require.NoError(t, err)
				val, err := q.Dequeue(ctx)
				require.NoError(t, err)
				require.Equal(t, 123, val)
				return q
			},
			wantVal:   234,
			timeout:   time.Second,
			wantData:  []int{0, 0, 0},
			wantSlice: []int{},
			wantLen:   0,
			wantHead:  2,
			wantTail:  2,
		},
		{
			name: "dequeue and empty last",
			q: func() *ArrayBlockingQueue[int] {
				q := NewArrayBlockingQueue[int](3)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := q.Enqueue(ctx, 123)
				require.NoError(t, err)
				err = q.Enqueue(ctx, 234)
				require.NoError(t, err)
				err = q.Enqueue(ctx, 345)
				require.NoError(t, err)
				val, err := q.Dequeue(ctx)
				require.NoError(t, err)
				require.Equal(t, 123, val)
				val, err = q.Dequeue(ctx)
				require.NoError(t, err)
				require.Equal(t, 234, val)
				return q
			},
			wantVal:   345,
			timeout:   time.Second,
			wantData:  []int{0, 0, 0},
			wantSlice: []int{},
			wantLen:   0,
			wantHead:  0,
			wantTail:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()
			q := tc.q()
			val, err := q.Dequeue(ctx)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantVal, val)
			assert.Equal(t, tc.wantData, q.data)
			assert.Equal(t, tc.wantSlice, q.AsSlice())
			assert.Equal(t, tc.wantLen, q.Len())
			assert.Equal(t, tc.wantHead, q.head)
			assert.Equal(t, tc.wantTail, q.tail)
		})
	}

	t.Run("dequeue timeout", func(t *testing.T) {
		q := NewArrayBlockingQueue[int](3)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		val, err := q.Dequeue(ctx)
		require.Equal(t, context.DeadlineExceeded, err)
		require.Equal(t, 0, val)
	})

	// 出队阻塞，然后入队，然后出队成功
	t.Run("dequeue blocking and enqueue", func(t *testing.T) {
		q := NewArrayBlockingQueue[int](3)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		go func() {
			time.Sleep(time.Millisecond * 100)
			err := q.Enqueue(ctx, 123)
			require.NoError(t, err)
		}()
		val, err := q.Dequeue(ctx)
		require.NoError(t, err)
		require.Equal(t, 123, val)
	})
}

func TestArrayBlockingQueue(t *testing.T) {
	// 并发测试，只是测试有没有死锁之类的问题
	// 先进先出这个特性依赖于其它单元测试
	// 也依赖于代码审查
	q := NewArrayBlockingQueue[int](100)
	var wg sync.WaitGroup
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			val := rand.Int()
			err := q.Enqueue(ctx, val)
			cancel()
			require.NoError(t, err)
		}()
	}
	go func() {
		for i := 0; i < 1000; i++ {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				_, err := q.Dequeue(ctx)
				cancel()
				require.NoError(t, err)
				wg.Done()
			}()
		}
	}()
	wg.Wait()
}

// ExampleNewArrayBlockingQueue
func ExampleNewArrayBlockingQueue() {
	q := NewArrayBlockingQueue[int](10)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = q.Enqueue(ctx, 22)
	val, err := q.Dequeue(ctx)
	// 这是例子，实际中你不需要写得那么复杂
	switch err {
	case context.Canceled:
		// 有人主动取消了，即调用了 cancel 方法。在这个例子里不会出现这个情况
	case context.DeadlineExceeded:
		// 超时了
	case nil:
		fmt.Println(val)
	default:
		// 其它乱七八糟的
	}
	// Output:
	// 22
}
