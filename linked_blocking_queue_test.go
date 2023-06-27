package concurrent_queue

import (
	"context"
	"fmt"
	"time"
)

func ExampleNewLinkedBlockingQueue() {
	// 创建一个容量为 10 的有界并发阻塞队列，如果传入 0 或者负数，那么创建的是无界并发阻塞队列
	q := NewLinkedBlockingQueue[int](10)
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
