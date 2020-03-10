package limiter

import (
	"errors"
	"sync"
	"time"

	"github.com/chatex-com/rate-limiter/pkg/config"
)

var (
	ErrZeroRuleCount    = errors.New("rule.Capacity must be a positive value")
	ErrZeroRuleInterval = errors.New("rule.Interval must be a positive value")
)

type Quota struct {
	cfg     config.Quota
	times   []time.Time
	timesMu sync.RWMutex
}

func NewQuota(cfg config.Quota) (*Quota, error) {
	if cfg.Capacity == 0 {
		return nil, ErrZeroRuleCount
	}

	if cfg.Interval <= 0 {
		return nil, ErrZeroRuleInterval
	}

	r := &Quota{
		cfg: cfg,
	}

	return r, nil
}

func (r *Quota) Add(t time.Time) {
	r.timesMu.Lock()
	defer r.timesMu.Unlock()

	r.times = append(r.times, t)

	go func() {
		<-time.After(r.cfg.Interval)

		r.timesMu.Lock()
		defer r.timesMu.Unlock()

		r.times = r.times[1:]
	}()
}

func (r *Quota) GetFreeSlot() (time.Duration, bool) {
	if r.freeSlots() > 0 {
		return 0, true
	}

	r.timesMu.RLock()
	defer r.timesMu.RUnlock()

	wait := r.cfg.Interval - time.Since(r.times[0])

	return wait, false
}

func (r *Quota) freeSlots() int32 {
	r.timesMu.RLock()
	defer r.timesMu.RUnlock()

	active := len(r.times)

	return int32(r.cfg.Capacity) - int32(active)
}
