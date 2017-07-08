// Package contextutil implements some context utility functions.
package contextutil

import (
	"context"
	"sync"
	"time"
)

// MultiContext creates a new context that shares state from all the given
// contexts. Deadline() returns the smallest Deadline() among all of the given
// contexts. Done() returns a channel that is closed when any of the given
// context's Done channels are closed and Err() returns its associated Err().
// Value() returns the first non-nil value from the given contexts, in the order
// given.
//
// Canceling this context releases resources associated with it, so code should
// call the returned cancel function as soon as the operations running in this
// Context complete.
func MultiContext(ctxs ...context.Context) (context.Context, context.CancelFunc) {
	mc := &multiContext{
		ctxs: ctxs,
		done: make(chan struct{}),
	}
	mc.selectCtxs()
	return mc, mc.cancel
}

type multiContext struct {
	ctxs []context.Context
	once sync.Once
	done chan struct{}
	err  error
}

func (mc *multiContext) cancel() {
	mc.once.Do(func() {
		close(mc.done)
		if mc.err == nil {
			mc.err = context.Canceled
		}
	})
}

func (mc *multiContext) selectCtxs() {
	for _, ctx := range mc.ctxs {
		go func(ctx context.Context) {
			select {
			case <-ctx.Done():
				mc.err = ctx.Err()
				mc.cancel()
			case <-mc.done:
			}
		}(ctx)
	}
}

func (mc *multiContext) Deadline() (deadline time.Time, ok bool) {
	var found bool
	min := time.Unix(1<<63-62135596801, 999999999)
	for _, ctx := range mc.ctxs {
		d, ok := ctx.Deadline()
		if ok {
			found = true
			if d.Before(min) {
				min = d
			}
		}
	}
	return min, found
}

func (mc *multiContext) Done() <-chan struct{} {
	return mc.done
}

func (mc *multiContext) Err() error {
	return mc.err
}

func (mc *multiContext) Value(key interface{}) interface{} {
	for _, ctx := range mc.ctxs {
		if v := ctx.Value(key); v != nil {
			return v
		}
	}
	return nil
}
