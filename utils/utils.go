package utils

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var (
	ErrDeadlineExceeded = errors.New("deadline exceeded")
)

type TimeoutRetryer struct {
	f                   func() (interface{}, error)
	DeadlineTimeout     time.Duration
	AttemptTimeout      time.Duration
	AttemptErrorHanlder func(err error)
}

func (r *TimeoutRetryer) Do() (interface{}, error) {
	ch := make(chan interface{})
	ctx, _ := context.WithTimeout(context.Background(), r.DeadlineTimeout)

	retrying := int32(1)
	go func() {
		for atomic.LoadInt32(&retrying) == 1 {
			c, err := r.f()
			if err != nil {
				if r.AttemptErrorHanlder != nil {
					r.AttemptErrorHanlder(err)
				}
				time.Sleep(r.AttemptTimeout)
				continue
			} else {
				ch <- c
				return
			}
		}
	}()

	select {
	case value := <-ch:
		return value, nil
	case <-ctx.Done():
		atomic.StoreInt32(&retrying, 0)
		return nil, ErrDeadlineExceeded
	}
}

func AwaitConnection(dialer func() (interface{}, error), timeout time.Duration) (interface{}, error) {
	return NewRetryer(dialer, timeout).Do()
}

func NewRetryer(f func() (interface{}, error), deadlineTimeout time.Duration) *TimeoutRetryer {
	return &TimeoutRetryer{
		f:                   f,
		DeadlineTimeout:     deadlineTimeout,
		AttemptTimeout:      500 * time.Millisecond,
		AttemptErrorHanlder: nil,
	}
}
