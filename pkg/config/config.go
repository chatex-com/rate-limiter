package config

import (
	"sync"
)

const (
	defaultConcurrency = 100
)

type Config struct {
	Concurrency uint32

	quotas   []*Quota
	quotasMu sync.RWMutex
}

func NewConfig() *Config {
	return &Config{
		Concurrency: defaultConcurrency,
	}
}

func NewConfigWithQuotas(quotas []*Quota) *Config {
	cfg := NewConfig()

	for _, rule := range quotas {
		cfg.AddQuota(rule)
	}

	return cfg
}

func (c *Config) AddQuota(quota *Quota) {
	c.quotasMu.Lock()
	defer c.quotasMu.Unlock()

	c.quotas = append(c.quotas, quota)
}

func (c *Config) GetQuotas() []Quota {
	c.quotasMu.Lock()
	defer c.quotasMu.Unlock()

	quotas := make([]Quota, len(c.quotas))
	for i, q := range c.quotas {
		quotas[i] = *q
	}

	return quotas
}
