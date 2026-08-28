package service

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
)

type syncJob struct {
	key  string
	run  func() error
	done chan struct{}
	err  error
}

type syncQueue struct {
	jobs    chan *syncJob
	mu      sync.Mutex
	pending map[string]*syncJob
}

func newSyncQueue(size int) *syncQueue {
	if size < 1 {
		size = 1
	}
	queue := &syncQueue{
		jobs:    make(chan *syncJob, size),
		pending: make(map[string]*syncJob),
	}
	go queue.run()
	return queue
}

func (q *syncQueue) run() {
	for job := range q.jobs {
		job.err = runSyncJob(job.run)
		q.mu.Lock()
		delete(q.pending, job.key)
		close(job.done)
		q.mu.Unlock()
	}
}

func runSyncJob(run func() error) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("同步工作器发生异常: %v", recovered)
		}
	}()
	return run()
}

func (q *syncQueue) do(key string, run func() error) error {
	q.mu.Lock()
	job, exists := q.pending[key]
	if !exists {
		job = &syncJob{key: key, run: run, done: make(chan struct{})}
		q.pending[key] = job
	}
	q.mu.Unlock()

	if !exists {
		q.jobs <- job
	}
	<-job.done
	return job.err
}

func syncQueueSize() int {
	size, err := strconv.Atoi(os.Getenv("SYNC_QUEUE_SIZE"))
	if err != nil || size <= 0 {
		return 100
	}
	return size
}

var globalSyncQueue = func() *syncQueue {
	size := syncQueueSize()
	log.Printf("微信同步队列已启动: workers=1 queue=%d", size)
	return newSyncQueue(size)
}()
