package bandwish_limiter

import (
	"time"
)

type ConfigRule struct {
	Count  uint
	Period time.Duration
}

func NewRule(count uint, period time.Duration) *ConfigRule {
	// FIXME: Check period and count for positive values

	return &ConfigRule{
		Count:  count,
		Period: period,
	}
}

func (r ConfigRule) getInterval() time.Duration {
	i := r.Period / time.Duration(r.Count)

	if i < minimalInterval {
		return minimalInterval
	}

	return i
}
