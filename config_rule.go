package job_runner

import (
	"time"
)

type ConfigRule struct {
	Count  uint
	Period time.Duration
}

func NewConfigRule(count uint, period time.Duration) *ConfigRule {
	return &ConfigRule{
		Count:  count,
		Period: period,
	}
}

func (r ConfigRule) getInterval() time.Duration {
	i := r.Period / time.Duration(r.Count)

	if i < minimalTickInterval {
		return minimalTickInterval
	}

	return i
}
