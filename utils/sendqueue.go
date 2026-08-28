package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
)

type sendJob struct {
	run    func() error
	result chan error
}

type serialSendQueue struct {
	jobs chan sendJob
}

func newSerialSendQueue(size int) *serialSendQueue {
	if size < 1 {
		size = 1
	}
	queue := &serialSendQueue{jobs: make(chan sendJob, size)}
	go func() {
		for job := range queue.jobs {
			job.result <- runSendJob(job.run)
			close(job.result)
		}
	}()
	return queue
}

func runSendJob(run func() error) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("发送工作器发生异常: %v", recovered)
		}
	}()
	return run()
}

func (q *serialSendQueue) do(run func() error) error {
	job := sendJob{run: run, result: make(chan error, 1)}
	q.jobs <- job
	return <-job.result
}

var (
	globalSendQueue     *serialSendQueue
	globalSendQueueOnce sync.Once
)

func enqueueSend(run func() error) error {
	globalSendQueueOnce.Do(func() {
		size, err := strconv.Atoi(os.Getenv("SEND_QUEUE_SIZE"))
		if err != nil || size <= 0 {
			size = 100
		}
		globalSendQueue = newSerialSendQueue(size)
		log.Printf("微信发送队列已启动: workers=1 queue=%d", size)
	})
	return globalSendQueue.do(run)
}
