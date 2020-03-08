package limiter

import (
	"sync"
	"time"

	"rate_limiter/pkg/config"
)

type QuotaGroup struct {
	quotas     []*Quota
	quotasLock sync.RWMutex
	lock       sync.Locker
}

func NewQuotaGroup(quotas []config.Quota) (*QuotaGroup, error) {
	list, err := createList(quotas)

	if err != nil {
		return nil, err
	}

	group := &QuotaGroup{
		quotas: list,
		lock:   &sync.Mutex{},
	}

	return group, nil
}

// Make a reservation for new slot. It means that you will
// immediately use it for execution query
//
// This method returns next values:
// 1st value - result of reservation. True = success, false = fail
// 2nd value - wait duration for next attempt if reservation was failed and zero otherwise
func (g *QuotaGroup) ReserveFreeSlot() (bool, time.Duration) {
	g.lock.Lock()
	defer g.lock.Unlock()

	if len(g.quotas) == 0 {
		return true, 0
	}

	var limited bool
	waits := make([]time.Duration, len(g.quotas))
	for _, q := range g.quotas {
		wait, free := q.GetFreeSlot()

		if !free {
			limited = true
			waits = append(waits, wait)
		}
	}

	if !limited {
		g.reserve()
		return true, 0
	}

	// find max duration from waits slice
	var wait time.Duration
	for _, w := range waits {
		if w > wait {
			wait = w
		}
	}

	return false, wait
}

func (g *QuotaGroup) reserve() {
	g.quotasLock.RLock()
	defer g.quotasLock.RUnlock()

	now := time.Now()
	for _, q := range g.quotas {
		q.Add(now)
	}
}

func createList(cfgQuotas []config.Quota) ([]*Quota, error) {
	quotas := make([]*Quota, len(cfgQuotas))

	for i, cfgQuota := range cfgQuotas {
		quota, err := NewQuota(cfgQuota)
		if err != nil {
			return nil, err
		}
		quotas[i] = quota
	}

	return quotas, nil
}
