package config

import (
	"time"
)

type Quota struct {
	Capacity uint
	Interval time.Duration
}

func NewQuota(capacity uint, interval time.Duration) *Quota {
	return &Quota{
		Capacity: capacity,
		Interval: interval,
	}
}
