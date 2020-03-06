package job_runner

import (
	"sync"
	"time"
)

const (
	// TODO: We should decide which minimal tick for request we will use (nano, micro, mini etc)
	minimalTickInterval     = time.Microsecond
	defaultConcurrencyLimit = 100
)

type Config struct {
	Strategy         Strategy
	ConcurrencyLimit uint32

	rules   []*ConfigRule
	rulesMu sync.RWMutex
}

func NewConfig() *Config {
	return &Config{
		Strategy:         StrategyImmediately,
		ConcurrencyLimit: defaultConcurrencyLimit,
	}
}

func NewConfigWithRules(rules []*ConfigRule) *Config {
	cfg := NewConfig()

	for _, rule := range rules {
		cfg.AddRule(rule)
	}

	return cfg
}

func (c *Config) AddRule(rule *ConfigRule) {
	c.rulesMu.Lock()
	defer c.rulesMu.Unlock()

	c.rules = append(c.rules, rule)
}

func (c Config) getTickInterval() time.Duration {
	var interval time.Duration

	switch c.Strategy {
	case StrategyImmediately:
		interval = minimalTickInterval
		break
	case StrategyEvenly:
		c.rulesMu.RLock()
		defer c.rulesMu.RUnlock()

		interval = c.calculateMinimalInterval()
	}

	return interval
}

func (c Config) calculateMinimalInterval() time.Duration {
	var interval time.Duration

	if len(c.rules) == 0 {
		return minimalTickInterval
	}

	for _, rule := range c.rules {
		i := rule.getInterval()

		if interval == 0 || i < interval {
			interval = i
		}
	}

	return interval
}
