package config

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewConfig(t *testing.T) {
	Convey("Empty settings (defaults)", t, func() {
		cfg := NewConfig()

		So(cfg.Concurrency, ShouldEqual, defaultConcurrency)
		So(cfg.quotas, ShouldBeEmpty)
	})

	Convey("Add rule", t, func() {
		cfg := NewConfig()
		rule := &Quota{}
		cfg.AddQuota(rule)

		So(cfg.quotas, ShouldContain, rule)
		So(cfg.quotas, ShouldHaveLength, 1)
	})
}

func TestNewConfigWithQuotas(t *testing.T) {
	Convey("Create config", t, func() {
		q1 := &Quota{}
		q2 := &Quota{}
		cfg := NewConfigWithQuotas([]*Quota{q1, q2})

		So(cfg.quotas, ShouldContain, q1)
		So(cfg.quotas, ShouldContain, q2)
		So(cfg.quotas, ShouldHaveLength, 2)
	})
}
