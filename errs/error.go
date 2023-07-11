package errs

import "errors"

var (
	ErrOutOfCapacity = errors.New("ekit: 超出最大容量限制")
	ErrEmptyQueue    = errors.New("ekit: 队列为空")
)
