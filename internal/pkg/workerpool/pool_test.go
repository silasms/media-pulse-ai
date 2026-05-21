package workerpool_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/silasms/media-pulse-ai/internal/pkg/workerpool"
)

func TestPool_ExecuteTasks(t *testing.T) {
	var counter int64
	handler := func(ctx context.Context, task int) {
		atomic.AddInt64(&counter, int64(task))
	}

	pool := workerpool.New[int](4, 100, handler)
	pool.Start()

	for i := 1; i <= 10; i++ {
		ok := pool.Submit(i)
		if !ok {
			t.Fatalf("failed to submit task %d", i)
		}
	}

	time.Sleep(50 * time.Millisecond)
	pool.Stop()

	if atomic.LoadInt64(&counter) != 55 {
		t.Fatalf("expected sum 55, got %d", atomic.LoadInt64(&counter))
	}
}

func BenchmarkWorkerPool(b *testing.B) {
	handler := func(ctx context.Context, task int) {}
	pool := workerpool.New[int](8, b.N, handler)
	pool.Start()
	defer pool.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Submit(i)
	}
}
