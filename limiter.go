package job_runner

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrJobExpired = errors.New("job was expired")
)

type Stat struct {
	inProgress uint32
	queue      uint32
	done       uint32
}

type Limiter struct {
	ticker   <-chan time.Time
	requests chan JobRequest
	wg       sync.WaitGroup
	stat     Stat

	rules []*limiterRule
	mu    sync.RWMutex
}

func NewLimiter(cfg *Config) *Limiter {
	limiter := &Limiter{
		requests: make(chan JobRequest, cfg.ConcurrencyLimit), // FIXME: Check it
		ticker:   cfg.getTicker(),
		rules:    make([]*limiterRule, len(cfg.rules)),
	}

	limiter.initRules(cfg.rules)

	go limiter.start()

	return limiter
}

func (l *Limiter) initRules(rules []ConfigRule) error {
	for i, rule := range rules {
		r, err := newLimiterRule(rule)

		if err != nil {
			return err
		}

		l.rules[i] = r
	}

	return nil
}

func (l *Limiter) Execute(job Job) <-chan JobResponse {
	return l.ExecuteWithExpiration(job, 0)
}

func (l *Limiter) ExecuteWithExpiration(job Job, timout time.Duration) <-chan JobResponse {
	return l.execute(job, timout)
}

func (l *Limiter) execute(job Job, timeout time.Duration) <-chan JobResponse {
	ch := make(chan JobResponse)

	l.wg.Add(1)

	r := JobRequest{
		job: job,
		ch:  ch,
	}

	if timeout > 0 {
		r.expiredAt = time.Now().Add(timeout)
	}

	// add request to the queue channel in separated goroutine
	// because the channel can be overload
	go func(r JobRequest) {
		l.requests <- r
	}(r)

	return ch
}

func (l *Limiter) Stat() Stat {
	return Stat{
		inProgress: atomic.LoadUint32(&l.stat.inProgress),
		queue:      atomic.LoadUint32(&l.stat.queue),
		done:       atomic.LoadUint32(&l.stat.done),
	}
}

func (l Limiter) AwaitAll() {
	l.wg.Wait()
}

func (l *Limiter) start() {
	for range l.ticker {
		// FIXME [√]: Implement the main limitations
		// FIXME [√]: Execution Strategy ("immediately" or "evenly")
		// FIXME [√]: Concurrency limitation - check it
		// FIXME [√]: Support timeouts in request

		if !l.hasConcurrentSlot() {
			continue
		}

		// Wait when current limitation will finish
		for {
			wait, free := l.getFreeSlot()

			if free {
				break
			}

			<-time.Tick(wait)
		}

		var req JobRequest
		for {
			req = <-l.requests

			if req.isExpired() {
				req.ch <- JobResponse{
					Result: nil,
					Error:  ErrJobExpired,
				}
				continue
			}

			break
		}

		go l.executeRequest(&req)
	}
}

func (l *Limiter) getFreeSlot() (time.Duration, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var limited bool

	waits := make([]time.Duration, len(l.rules))
	for _, r := range l.rules {
		d, free := r.getFreeSlot()

		if !free {
			limited = true
			waits = append(waits, d)
		}
	}

	if !limited {
		return 0, true
	}

	// find max duration from waits slice
	var wait time.Duration
	for _, w := range waits {
		if w > wait {
			wait = w
		}
	}

	return wait, false
}

func (l Limiter) hasConcurrentSlot() bool {
	inProgress := atomic.LoadUint32(&l.stat.inProgress)
	limit := uint32(cap(l.requests))

	return limit > inProgress
}

func (l *Limiter) executeRequest(r *JobRequest) {
	l.mu.Lock()
	now := time.Now()
	for _, r := range l.rules {
		r.add(now)
	}
	l.mu.Unlock()

	result, err := r.job()

	atomic.AddUint32(&l.stat.done, 1)
	atomic.AddUint32(&l.stat.inProgress, -1)

	r.ch <- JobResponse{
		Result: result,
		Error:  err,
	}
	close(r.ch)

	l.wg.Done()
}
