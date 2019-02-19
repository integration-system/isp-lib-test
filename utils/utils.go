package utils

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var (
	ErrDeadlineExceeded = errors.New("connect deadline exceeded")
)

func AwaitConnection(dialer func() (interface{}, error), timeout time.Duration) (interface{}, error) {
	connChan := make(chan interface{})
	ctx, _ := context.WithTimeout(context.Background(), timeout)

	reconnect := int32(1)
	go func() {
		for atomic.LoadInt32(&reconnect) == 1 {
			c, err := dialer()
			if err != nil {
				time.Sleep(500 * time.Millisecond)
				continue
			} else {
				connChan <- c
				return
			}
		}
	}()

	select {
	case conn := <-connChan:
		return conn, nil
	case <-ctx.Done():
		atomic.StoreInt32(&reconnect, 0)
		return nil, ErrDeadlineExceeded
	}
}
