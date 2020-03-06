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
	inProgress int32
	queue      int32
	expired    int32
	done       int32
}

type Runner struct {
	ticker   <-chan time.Time
	requests chan JobRequest
	wg       sync.WaitGroup
	stat     Stat

	rules []*runnerRule
	mu    sync.RWMutex
}

func NewRunner(cfg *Config) (*Runner, error) {
	interval := cfg.getTickInterval()

	runner := &Runner{
		requests: make(chan JobRequest, cfg.ConcurrencyLimit), // FIXME: Check it
		ticker:   time.Tick(interval),
		rules:    make([]*runnerRule, len(cfg.rules)),
	}

	err := runner.initRules(cfg.rules)
	if err != nil {
		return nil, err
	}

	go runner.loop()

	return runner, nil
}

func (l *Runner) initRules(rules []*ConfigRule) error {
	for i, rule := range rules {
		r, err := newRunnerRule(rule)

		if err != nil {
			return err
		}

		l.rules[i] = r
	}

	return nil
}

func (l *Runner) Execute(job Job) <-chan JobResponse {
	return l.ExecuteWithExpiration(job, 0)
}

func (l *Runner) ExecuteWithExpiration(job Job, timout time.Duration) <-chan JobResponse {
	return l.execute(job, timout)
}

func (l *Runner) execute(job Job, timeout time.Duration) <-chan JobResponse {
	atomic.AddInt32(&l.stat.queue, 1)

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

func (l *Runner) Stat() Stat {
	return Stat{
		inProgress: atomic.LoadInt32(&l.stat.inProgress),
		queue:      atomic.LoadInt32(&l.stat.queue),
		done:       atomic.LoadInt32(&l.stat.done),
		expired:    atomic.LoadInt32(&l.stat.expired),
	}
}

func (l Runner) AwaitAll() {
	l.wg.Wait()
}

func (l *Runner) loop() {
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
				atomic.AddInt32(&l.stat.expired, 1)
				atomic.AddInt32(&l.stat.queue, -1)
				close(req.ch)
				l.wg.Done()
				continue
			}

			break
		}

		go l.executeRequest(&req)
	}
}

func (l *Runner) getFreeSlot() (time.Duration, bool) {
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

func (l Runner) hasConcurrentSlot() bool {
	inProgress := atomic.LoadInt32(&l.stat.inProgress)
	limit := int32(cap(l.requests))

	return limit > inProgress
}

func (l *Runner) executeRequest(r *JobRequest) {
	atomic.AddInt32(&l.stat.inProgress, 1)
	atomic.AddInt32(&l.stat.queue, -1)

	l.mu.Lock()
	now := time.Now()
	for _, r := range l.rules {
		r.add(now)
	}
	l.mu.Unlock()

	result, err := r.job()

	atomic.AddInt32(&l.stat.done, 1)
	atomic.AddInt32(&l.stat.inProgress, -1)

	r.ch <- JobResponse{
		Result: result,
		Error:  err,
	}
	close(r.ch)

	l.wg.Done()
}
