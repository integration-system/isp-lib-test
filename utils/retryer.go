package utils

import (
	"errors"
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
	deadLine := time.After(r.DeadlineTimeout)
	// use closed channel to trigger select on the first iteration of the loop
	tempCh := make(chan time.Time)
	close(tempCh)
	var retryCh <-chan time.Time = tempCh

	for {
		select {
		case <-deadLine:
			return nil, ErrDeadlineExceeded
		case <-retryCh:
		}

		c, err := r.f()
		if err != nil {
			if r.AttemptErrorHanlder != nil {
				r.AttemptErrorHanlder(err)
			}
			retryCh = time.After(r.AttemptTimeout)
			continue
		}

		return c, nil
	}
}

func AwaitConnection(dialer func() (interface{}, error), timeout time.Duration) (interface{}, error) {
	return NewRetryer(dialer, timeout).Do()
}

func NewRetryer(f func() (interface{}, error), deadlineTimeout time.Duration) *TimeoutRetryer {
	return &TimeoutRetryer{
		f:                   f,
		DeadlineTimeout:     deadlineTimeout,
		AttemptTimeout:      200 * time.Millisecond,
		AttemptErrorHanlder: nil,
	}
}
