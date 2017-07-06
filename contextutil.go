package contextutil

import (
	"context"
	"sync"
	"time"
)

// MultiContext creates a new context that is shares state from all the given
// contexts. Deadline() returns the smallest Deadline() among all of the given
// contexts. Done() returns a channel that is closed when any of the given
// context's Done channels are closed and Err() returns it associated Err().
// Value() returns the first non-nil value from the given contexts, in order the
// order given.
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

func (c *multiContext) cancel() {
	c.once.Do(func() {
		close(c.done)
		if c.err == nil {
			c.err = context.Canceled
		}
	})
}

func (c *multiContext) selectCtxs() {
	for _, ctx := range c.ctxs {
		go func(ctx context.Context) {
			select {
			case <-ctx.Done():
				c.err = ctx.Err()
				c.cancel()
			case <-c.done:
			}
		}(ctx)
	}
}

func (c *multiContext) Deadline() (deadline time.Time, ok bool) {
	var found bool
	min := time.Unix(1<<63-62135596801, 999999999)
	for _, ctx := range c.ctxs {
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

func (c *multiContext) Done() <-chan struct{} {
	return c.done
}

func (c *multiContext) Err() error {
	return c.err
}

func (c *multiContext) Value(key interface{}) interface{} {
	for _, ctx := range c.ctxs {
		if v := ctx.Value(key); v != nil {
			return v
		}
	}
	return nil
}
