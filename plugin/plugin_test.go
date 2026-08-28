package plugin

import (
	"github.com/loks666/FlyAI/model/dto"
	"sync"
	"testing"
	"time"
)

func TestAsyncDispatcherPreservesOrderWithinGroup(t *testing.T) {
	t.Setenv("PLUGIN_WORKERS", "2")
	t.Setenv("PLUGIN_QUEUE_SIZE", "20")
	dispatcher := NewMessageMatchDispatcher()
	dispatcher.SetAsync(true)

	var mu sync.Mutex
	order := make([]int64, 0, 10)
	done := make(chan struct{})
	dispatcher.RegisterHandler(func(*dto.Message) bool { return true }, func(ctx *MessageContext) {
		mu.Lock()
		order = append(order, ctx.MsgId)
		if len(order) == 10 {
			close(done)
		}
		mu.Unlock()
	})

	for id := int64(1); id <= 10; id++ {
		dispatcher.Dispatch(&dto.Message{MsgId: id, FromUser: "same-group@chatroom"})
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for plugin jobs")
	}

	mu.Lock()
	defer mu.Unlock()
	for index, id := range order {
		if want := int64(index + 1); id != want {
			t.Fatalf("message order[%d] = %d, want %d", index, id, want)
		}
	}
}

func TestAsyncDispatcherWorkerSurvivesPanic(t *testing.T) {
	t.Setenv("PLUGIN_WORKERS", "1")
	dispatcher := NewMessageMatchDispatcher()
	dispatcher.SetAsync(true)
	done := make(chan struct{})
	dispatcher.RegisterHandler(func(*dto.Message) bool { return true }, func(ctx *MessageContext) {
		if ctx.MsgId == 1 {
			panic("boom")
		}
		close(done)
	})
	dispatcher.Dispatch(&dto.Message{MsgId: 1, FromUser: "group"})
	dispatcher.Dispatch(&dto.Message{MsgId: 2, FromUser: "group"})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not process the message after a panic")
	}
}
