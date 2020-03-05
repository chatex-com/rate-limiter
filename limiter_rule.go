package bandwish_limiter

import (
	"sync"
	"time"
)

type limiterRule struct {
	cfg ConfigRule

	times   []time.Time
	timesMu sync.RWMutex
}

func newLimiterRule(rule ConfigRule) limiterRule {
	r := limiterRule{
		cfg: rule,
	}

	return r
}

func (r *limiterRule) add(t time.Time) {
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

func (r *limiterRule) getFreeSlot() (time.Duration, bool) {
	if r.freeSlots() > 0 {
		return 0, true
	}

	r.timesMu.RLock()
	defer r.timesMu.RUnlock()

	if len(r.times) == 0 {
		return 0, true
	}

	return time.Since(r.times[0]), false
}

func (r *limiterRule) freeSlots() int32 {
	r.timesMu.RLock()
	defer r.timesMu.RUnlock()

	active := len(r.times)

	return int32(r.cfg.Count) - int32(active)
}
