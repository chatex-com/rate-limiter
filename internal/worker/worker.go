package worker

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/chatex-com/rate-limiter/internal/limiter"
	"github.com/chatex-com/rate-limiter/pkg/job"
)

type Worker struct {
	quotas        *limiter.QuotaGroup
	requests      <-chan job.Request
	wg            *sync.WaitGroup
	isRunning     bool
	isRunningLock sync.RWMutex
	stat          Stat
}

func NewWorker(quotas *limiter.QuotaGroup, requests <-chan job.Request, wg *sync.WaitGroup) *Worker {
	return &Worker{
		quotas:   quotas,
		requests: requests,
		wg:       wg,
	}
}

func (w *Worker) Start() {
	w.isRunningLock.Lock()
	defer w.isRunningLock.Unlock()

	if w.isRunning {
		return
	}

	w.isRunning = true

	go w.loop()
}

func (w *Worker) Stop() {
	w.isRunningLock.Lock()
	defer w.isRunningLock.Unlock()

	if !w.isRunning {
		return
	}

	w.isRunning = false
}

func (w *Worker) IsRunning() bool {
	w.isRunningLock.RLock()
	defer w.isRunningLock.RUnlock()

	return w.isRunning
}

func (w *Worker) loop() {
	for w.IsRunning() {
		request := <-w.requests

		if request.IsExpired() {
			w.error(request, job.ErrJobExpired)
			continue
		}

		err := w.reserveFreeSlot(request)
		if err != nil {
			w.error(request, err)
			continue
		}

		w.execute(request)
	}
}

func (w *Worker) reserveFreeSlot(request job.Request) error {
	for {
		free, wait := w.quotas.ReserveFreeSlot()

		if free {
			return nil
		}

		if request.IsExpiredAfter(wait) {
			return job.ErrJobExpired
		}

		<-time.After(wait)
	}
}

func (w *Worker) execute(request job.Request) {
	atomic.AddInt64(&w.stat.InProcess, 1)
	result, err := request.Job()

	request.Ch <- job.Response{
		Result: result,
		Error:  err,
	}

	atomic.AddInt64(&w.stat.InProcess, -1)
	if err == nil {
		atomic.AddInt64(&w.stat.Done, 1)
	} else {
		atomic.AddInt64(&w.stat.Error, 1)
	}

	close(request.Ch)

	w.wg.Done()
}

func (w *Worker) error(request job.Request, err error) {
	request.Ch <- job.Response{
		Result: nil,
		Error:  err,
	}

	atomic.AddInt64(&w.stat.Error, 1)
	close(request.Ch)

	w.wg.Done()
}
