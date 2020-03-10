package limiter

import (
	"sync"
	"time"

	"github.com/chatex-com/rate-limiter/internal/limiter"
	"github.com/chatex-com/rate-limiter/internal/worker"
	"github.com/chatex-com/rate-limiter/pkg/config"
	"github.com/chatex-com/rate-limiter/pkg/job"
)

type RateLimiter struct {
	quotas        *limiter.QuotaGroup
	workers       []*worker.Worker
	requests      chan job.Request
	isRunning     bool
	isRunningLock sync.Locker
	wg            sync.WaitGroup
}

func NewRateLimiter(cfg *config.Config) (*RateLimiter, error) {
	quotas, err := limiter.NewQuotaGroup(cfg.GetQuotas())
	if err != nil {
		return nil, err
	}

	l := &RateLimiter{
		quotas:        quotas,
		isRunningLock: &sync.Mutex{},
		requests:      make(chan job.Request, cfg.Concurrency),
	}

	l.init(cfg.Concurrency)

	return l, nil
}

func (l *RateLimiter) init(concurrency uint32) {
	l.workers = make([]*worker.Worker, concurrency)

	for i := uint32(0); i < concurrency; i++ {
		l.workers[i] = worker.NewWorker(l.quotas, l.requests, &l.wg)
	}
}

func (l *RateLimiter) Execute(j job.Job) <-chan job.Response {
	return l.ExecuteWithTimout(j, 0)
}

func (l *RateLimiter) ExecuteWithTimout(j job.Job, timeout time.Duration) <-chan job.Response {
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

	l.isRunning = true
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

	l.isRunning = false
}

func (l *RateLimiter) AwaitAll() {
	l.wg.Wait()
}
