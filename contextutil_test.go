package contextutil

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestMultiContextSingle(t *testing.T) {
	mc, cancel := MultiContext(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		select {
		case <-mc.Done():
		case <-time.After(2 * time.Second):
			t.Errorf("done never closed")
		}
		wg.Done()
	}()
	cancel()
	wg.Wait()
	if mc.Err() != context.Canceled {
		t.Errorf("didn't get cancel errors: %v", mc.Err())
	}
}

func TestMultiContextCancel(t *testing.T) {
	mc, cancel := MultiContext(context.Background(), context.TODO())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		select {
		case <-mc.Done():
		case <-time.After(2 * time.Second):
			t.Errorf("done never closed")
		}
		wg.Done()
	}()
	cancel()
	wg.Wait()
	if mc.Err() != context.Canceled {
		t.Errorf("didn't get cancel errors: %v", mc.Err())
	}
}

func TestMultiContextContextCancel(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	mc, cancel := MultiContext(context.Background(), context.TODO(), ctx)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		select {
		case <-mc.Done():
		case <-time.After(2 * time.Second):
			t.Errorf("done never closed")
		}
		wg.Done()
	}()
	ctxCancel()
	wg.Wait()
	if mc.Err() != context.Canceled {
		t.Errorf("didn't get cancel errors: %v", mc.Err())
	}
}

func TestMultiContextContextDeadline(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 200*time.Microsecond)
	defer ctxCancel()
	mc, cancel := MultiContext(context.Background(), context.TODO(), ctx)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		select {
		case <-mc.Done():
		case <-time.After(2 * time.Second):
			t.Errorf("done never closed")
		}
		wg.Done()
	}()
	wg.Wait()
	if mc.Err() != context.DeadlineExceeded {
		t.Errorf("didn't get cancel errors: %v", mc.Err())
	}
}

func TestMultiContextDeadline(t *testing.T) {
	now := time.Now().Add(1 * time.Hour)
	var ctxs []context.Context
	for x := 0; x < 10; x++ {
		ctx, cancel := context.WithDeadline(context.Background(), now.Add(time.Duration(1000+x)*time.Second))
		defer cancel()
		ctxs = append(ctxs, ctx)
	}
	ctx, cancel := MultiContext(ctxs...)
	defer cancel()
	dl, ok := ctx.Deadline()
	if !ok || dl != now.Add(time.Duration(1000+0)*time.Second) {
		t.Errorf("expected deadline %v, got %v", now.Add(time.Duration(1000+0)*time.Second), dl)
	}
}

func TestMultiContextNoDeadline(t *testing.T) {
	var ctxs []context.Context
	for x := 0; x < 10; x++ {
		ctxs = append(ctxs, context.Background())
	}
	ctx, cancel := MultiContext(ctxs...)
	defer cancel()
	dl, ok := ctx.Deadline()
	if ok {
		t.Errorf("expected zero deadline, got %v", dl)
	}
}

type Key int

var (
	one Key = 1
)

func TestMultiContextValue(t *testing.T) {
	ctx1 := context.WithValue(context.Background(), one, 1)
	ctx2 := context.WithValue(context.Background(), one, 2)
	ctx, cancel := MultiContext(ctx1, ctx2)
	defer cancel()
	if i, ok := ctx.Value(one).(int); !ok || i != 1 {
		t.Errorf("expected value %v, got %v", 1, i)
	}
	if i, ok := ctx.Value(3).(int); ok {
		t.Errorf("expected value %v; got %v", nil, i)
	}
}
