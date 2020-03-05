package job_runner

import (
	"sync"
	"time"
)

const (
	// TODO: We should decide which minimal tick for request we will use (nano, micro, mini etc)
	minimalInterval         = time.Microsecond
	defaultConcurrencyLimit = 100
)

type Config struct {
	Strategy         Strategy
	ConcurrencyLimit uint32

	rules   []ConfigRule
	rulesMu sync.RWMutex
}

func NewConfig() *Config {
	return &Config{
		Strategy:         StrategyImmediately,
		ConcurrencyLimit: defaultConcurrencyLimit,
	}
}

func NewConfigWithRules(rules []ConfigRule) *Config {
	cfg := NewConfig()

	for _, rule := range rules {
		cfg.AddRule(rule)
	}

	return cfg
}

func (c *Config) AddRule(rule ConfigRule) {
	c.rulesMu.Lock()
	defer c.rulesMu.Unlock()

	c.rules = append(c.rules, rule)
}

func (c Config) getTicker() <-chan time.Time {
	var interval time.Duration
	switch c.Strategy {
	case StrategyImmediately:
		interval = minimalInterval
		break
	case StrategyEvenly:
	default:
		c.rulesMu.RLock()
		defer c.rulesMu.RUnlock()

		interval = calculateMinimalInterval(c.rules)
	}

	return time.Tick(interval)
}

func calculateMinimalInterval(rules []ConfigRule) time.Duration {
	var interval time.Duration

	if len(rules) == 0 {
		return minimalInterval
	}

	for _, rule := range rules {
		i := rule.getInterval()

		if interval == 0 || i < interval {
			interval = i
		}
	}

	return interval
}
