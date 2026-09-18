package singleflight

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestDoSharesConcurrentResult(t *testing.T) {
	group := NewGroup()
	const callers = 8

	var mu sync.Mutex
	started := make(chan struct{})
	release := make(chan struct{})
	startedCalls := 0
	results := make(chan interface{}, callers)
	errors := make(chan error, callers)
	var wg sync.WaitGroup
	fn := func() (interface{}, error) {
		mu.Lock()
		startedCalls++
		close(started)
		mu.Unlock()
		<-release
		return "shared-value", nil
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		value, err := group.Do("shared-key", fn)
		results <- value
		errors <- err
	}()

	<-started
	for i := 1; i < callers; i++ {
		waiterDone := make(chan struct{})
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := group.Do("shared-key", fn)
			results <- value
			errors <- err
			close(waiterDone)
		}()
		select {
		case <-waiterDone:
			t.Fatal("waiter completed before the original call was released")
		case <-time.After(10 * time.Millisecond):
		}
	}
	close(release)
	wg.Wait()
	close(results)
	close(errors)

	if startedCalls != 1 {
		t.Fatalf("expected function to run once, got %d calls", startedCalls)
	}
	for value := range results {
		if value != "shared-value" {
			t.Fatalf("unexpected result: %v", value)
		}
	}
	for err := range errors {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestDoSharesErrorsAndAllowsRetry(t *testing.T) {
	group := NewGroup()
	wantErr := errors.New("backend unavailable")

	value, err := group.Do("error-key", func() (interface{}, error) {
		return nil, wantErr
	})
	if value != nil || !errors.Is(err, wantErr) {
		t.Fatalf("expected shared error, got value=%v err=%v", value, err)
	}

	value, err = group.Do("error-key", func() (interface{}, error) {
		return "retry-value", nil
	})
	if err != nil || value != "retry-value" {
		t.Fatalf("expected retry to execute, got value=%v err=%v", value, err)
	}
}
