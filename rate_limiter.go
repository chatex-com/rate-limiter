package rate_limiter

import (
	"sync"
	"time"

	"rate_limiter/internal/limiter"
	"rate_limiter/internal/worker"
	"rate_limiter/pkg/config"
	"rate_limiter/pkg/job"
)

type RateLimiter struct {
	concurrency   uint32
	quotas        *limiter.QuotaGroup
	workers       []*worker.Worker
	requests      chan job.Request
	isRunning     bool
	isRunningLock sync.Locker
	wg            sync.WaitGroup
}

func NewRateLimiter(cfg config.Config) (*RateLimiter, error) {
	quotas, err := limiter.NewQuotaGroup(cfg.GetQuotas())
	if err != nil {
		return nil, err
	}

	l := &RateLimiter{
		concurrency:   cfg.Concurrency,
		quotas:        quotas,
		isRunningLock: &sync.Mutex{},
		requests:      make(chan job.Request, cfg.Concurrency),
	}

	l.init()

	return l, nil
}

func (l *RateLimiter) init() {
	l.workers = make([]*worker.Worker, l.concurrency)

	done := make(chan bool, l.concurrency)
	for i := uint32(0); i < l.concurrency; i++ {
		l.workers[i] = worker.NewWorker(l.quotas, l.requests, done)
	}

	go func(done chan bool) {
		for range done {
			l.wg.Done()
		}
	}(done)
}

func (l *RateLimiter) Execute(job job.Job) <-chan job.Response {
	return l.ExecuteWithTimout(job, 0)
}

func (l *RateLimiter) ExecuteWithTimout(j job.Job, timeout time.Duration) <-chan job.Response {
	// atomic.AddInt32(&l.stat.queue, 1)
	l.wg.Add(1)

	ch := make(chan job.Response)

	r := job.Request{
		Job: j,
		Ch:  ch,
	}

	if timeout > 0 {
		r.ExpiredAt = time.Now().Add(timeout)
	}

	// add request to the queue channel in separated goroutine
	// because the channel can be overload
	go func(r job.Request) {
		l.requests <- r
	}(r)

	return ch
}

func (l *RateLimiter) Start() {
	l.isRunningLock.Lock()
	defer l.isRunningLock.Unlock()

	if l.isRunning {
		return
	}

	for _, w := range l.workers {
		w.Start()
	}
}

func (l *RateLimiter) Stop() {
	l.isRunningLock.Lock()
	defer l.isRunningLock.Unlock()

	if !l.isRunning {
		return
	}

	for _, w := range l.workers {
		w.Stop()
	}
}
func (l *RateLimiter) AwaitAll() {
	l.wg.Wait()
}
