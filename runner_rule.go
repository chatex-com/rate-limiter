package job_runner

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrZeroRuleCount  = errors.New("rule.Count must be a positive value")
	ErrZeroRulePeriod = errors.New("rule.Period must be a positive value")
)

type runnerRule struct {
	cfg *ConfigRule

	times   []time.Time
	timesMu sync.RWMutex
}

func newRunnerRule(rule *ConfigRule) (*runnerRule, error) {
	if rule.Count <= 0 {
		return nil, ErrZeroRuleCount
	}

	if rule.Period <= 0 {
		return nil, ErrZeroRulePeriod
	}

	r := &runnerRule{
		cfg: rule,
	}

	return r, nil
}

func (r *runnerRule) add(t time.Time) {
	r.timesMu.Lock()
	defer r.timesMu.Unlock()

	r.times = append(r.times, t)

	go func() {
		<-time.After(r.cfg.Period)

		r.timesMu.Lock()
		defer r.timesMu.Unlock()

		r.times = r.times[1:]
	}()
}

func (r *runnerRule) getFreeSlot() (time.Duration, bool) {
	if r.freeSlots() > 0 {
		return 0, true
	}

	r.timesMu.RLock()
	defer r.timesMu.RUnlock()

	wait := r.cfg.Period - time.Since(r.times[0])

	return wait, false
}

func (r *runnerRule) freeSlots() int32 {
	r.timesMu.RLock()
	defer r.timesMu.RUnlock()

	active := len(r.times)

	return int32(r.cfg.Count) - int32(active)
}
