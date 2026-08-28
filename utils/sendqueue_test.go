package utils

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSerialSendQueueRunsOneJobAtATime(t *testing.T) {
	queue := newSerialSendQueue(32)
	var active int32
	var maxActive int32
	var completed int32
	var callers sync.WaitGroup

	for i := 0; i < 20; i++ {
		callers.Add(1)
		go func() {
			defer callers.Done()
			if err := queue.do(func() error {
				current := atomic.AddInt32(&active, 1)
				for {
					maximum := atomic.LoadInt32(&maxActive)
					if current <= maximum || atomic.CompareAndSwapInt32(&maxActive, maximum, current) {
						break
					}
				}
				time.Sleep(time.Millisecond)
				atomic.AddInt32(&active, -1)
				atomic.AddInt32(&completed, 1)
				return nil
			}); err != nil {
				t.Errorf("send job failed: %v", err)
			}
		}()
	}
	callers.Wait()

	if got := atomic.LoadInt32(&maxActive); got != 1 {
		t.Fatalf("maximum concurrent send jobs = %d, want 1", got)
	}
	if got := atomic.LoadInt32(&completed); got != 20 {
		t.Fatalf("completed send jobs = %d, want 20", got)
	}
}

func TestSerialSendQueueSurvivesPanickingJob(t *testing.T) {
	queue := newSerialSendQueue(2)
	if err := queue.do(func() error { panic("boom") }); err == nil {
		t.Fatal("panicking send job should return an error")
	}
	want := errors.New("next job")
	if err := queue.do(func() error { return want }); !errors.Is(err, want) {
		t.Fatalf("worker did not process the next job: %v", err)
	}
}
