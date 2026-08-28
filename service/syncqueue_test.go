package service

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSyncQueueDeduplicatesPendingKey(t *testing.T) {
	queue := newSyncQueue(20)
	started := make(chan struct{})
	release := make(chan struct{})
	var runs int32
	var callers sync.WaitGroup

	run := func() error {
		if atomic.AddInt32(&runs, 1) == 1 {
			close(started)
		}
		<-release
		return nil
	}
	for i := 0; i < 10; i++ {
		callers.Add(1)
		go func() {
			defer callers.Done()
			if err := queue.do("group:one", run); err != nil {
				t.Errorf("sync job failed: %v", err)
			}
		}()
	}
	<-started
	time.Sleep(10 * time.Millisecond)
	close(release)
	callers.Wait()

	if got := atomic.LoadInt32(&runs); got != 1 {
		t.Fatalf("sync job ran %d times, want 1", got)
	}
}

func TestSyncQueueRunsDifferentKeysSerially(t *testing.T) {
	queue := newSyncQueue(10)
	var active int32
	var maxActive int32
	var callers sync.WaitGroup

	for _, key := range []string{"one", "two", "three"} {
		key := key
		callers.Add(1)
		go func() {
			defer callers.Done()
			_ = queue.do(key, func() error {
				current := atomic.AddInt32(&active, 1)
				for {
					maximum := atomic.LoadInt32(&maxActive)
					if current <= maximum || atomic.CompareAndSwapInt32(&maxActive, maximum, current) {
						break
					}
				}
				time.Sleep(time.Millisecond)
				atomic.AddInt32(&active, -1)
				return nil
			})
		}()
	}
	callers.Wait()

	if got := atomic.LoadInt32(&maxActive); got != 1 {
		t.Fatalf("maximum concurrent sync jobs = %d, want 1", got)
	}
}
