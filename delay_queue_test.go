package concurrent_queue

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestDelayQueue_Enqueue(t *testing.T) {
	t.Parallel()
	now := time.Now()
	testCases := []struct {
		name    string
		q       *DelayQueue[delayElem]
		timeout time.Duration
		val     delayElem
		wantErr error
	}{
		{
			name:    "enqueued",
			q:       NewDelayQueue[delayElem](3),
			timeout: time.Second,
			val:     delayElem{val: 123, deadline: now.Add(time.Minute)},
		},
		{
			// context 本身已经过期了
			name:    "invalid context",
			q:       NewDelayQueue[delayElem](3),
			timeout: -time.Second,
			val:     delayElem{val: 123, deadline: now.Add(time.Minute)},
			wantErr: context.DeadlineExceeded,
		},
		{
			// enqueue 的时候阻塞住了，直到超时
			name:    "enqueue timeout",
			q:       newDelayQueue(t, delayElem{val: 123, deadline: now.Add(time.Minute)}),
			timeout: time.Millisecond * 100,
			val:     delayElem{val: 234, deadline: now.Add(time.Minute)},
			wantErr: context.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()
			err := tc.q.EnQueue(ctx, tc.val)
			assert.Equal(t, tc.wantErr, err)
		})
	}

	// 队列满了，这时候入队。
	// 在等待一段时间之后，队列元素被取走一个
	t.Run("enqueue while DeQueue", func(t *testing.T) {
		t.Parallel()
		q := newDelayQueue(t, delayElem{val: 123, deadline: time.Now().Add(time.Second)})
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()
			ele, err := q.DeQueue(ctx)
			require.NoError(t, err)
			require.Equal(t, 123, ele.val)
		}()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		err := q.EnQueue(ctx, delayElem{val: 345, deadline: time.Now().Add(time.Millisecond * 1500)})
		require.NoError(t, err)
	})

	// 入队相同过期时间的元素
	// 但是因为我们在入队的时候是分别计算 Delay 的
	// 那么就会导致虽然过期时间是相同的，但是因为调用 Delay 有先后之分
	// 所以会造成 dstDelay 就是要比 srcDelay 小一点
	t.Run("enqueue with same deadline", func(t *testing.T) {
		t.Parallel()
		q := NewDelayQueue[delayElem](3)
		deadline := time.Now().Add(time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		err := q.EnQueue(ctx, delayElem{val: 123, deadline: deadline})
		require.NoError(t, err)
		err = q.EnQueue(ctx, delayElem{val: 456, deadline: deadline})
		require.NoError(t, err)
		err = q.EnQueue(ctx, delayElem{val: 789, deadline: deadline})
		require.NoError(t, err)

		ele, err := q.DeQueue(ctx)
		require.NoError(t, err)
		require.Equal(t, 123, ele.val)

		ele, err = q.DeQueue(ctx)
		require.NoError(t, err)
		require.Equal(t, 789, ele.val)

		ele, err = q.DeQueue(ctx)
		require.NoError(t, err)
		require.Equal(t, 456, ele.val)
	})
}

func newDelayQueue(t *testing.T, eles ...delayElem) *DelayQueue[delayElem] {
	q := NewDelayQueue[delayElem](len(eles))
	for _, ele := range eles {
		err := q.EnQueue(context.Background(), ele)
		require.NoError(t, err)
	}
	return q
}

type delayElem struct {
	deadline time.Time
	val      int
}

func (d delayElem) Delay() time.Duration {
	return time.Until(d.deadline)
}
