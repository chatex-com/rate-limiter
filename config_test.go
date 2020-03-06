package job_runner

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewConfig(t *testing.T) {
	Convey("Empty settings (defaults)", t, func() {
		cfg := NewConfig()

		So(cfg.Strategy, ShouldEqual, StrategyImmediately)
		So(cfg.ConcurrencyLimit, ShouldEqual, defaultConcurrencyLimit)
		So(cfg.rules, ShouldBeEmpty)
	})

	Convey("Add rule", t, func() {
		cfg := NewConfig()
		rule := &ConfigRule{}
		cfg.AddRule(rule)

		So(cfg.rules, ShouldContain, rule)
		So(cfg.rules, ShouldHaveLength, 1)
	})
}

func TestNewConfigWithRules(t *testing.T) {
	Convey("Create config", t, func() {
		rule1 := &ConfigRule{}
		rule2 := &ConfigRule{}
		cfg := NewConfigWithRules([]*ConfigRule{rule1, rule2})

		So(cfg.rules, ShouldContain, rule1)
		So(cfg.rules, ShouldContain, rule2)
		So(cfg.rules, ShouldHaveLength, 2)
	})
}

func TestGetTicker(t *testing.T) {
	Convey("StrategyImmediately", t, func() {
		cfg := NewConfig()
		cfg.Strategy = StrategyImmediately

		So(cfg.getTickInterval(), ShouldEqual, minimalTickInterval)
	})

	Convey("StrategyEvenly without rules", t, func() {
		cfg := NewConfig()
		cfg.Strategy = StrategyEvenly

		So(cfg.getTickInterval(), ShouldEqual, minimalTickInterval)
	})

	Convey("StrategyEvenly with one rule", t, func() {
		cfg := NewConfig()
		cfg.Strategy = StrategyEvenly
		cfg.AddRule(NewConfigRule(10, time.Minute))

		So(cfg.getTickInterval(), ShouldEqual, 6 * time.Second)
	})

	Convey("StrategyEvenly with multiple rules", t, func() {
		cfg := NewConfig()
		cfg.Strategy = StrategyEvenly
		cfg.AddRule(NewConfigRule(10, time.Second))
		cfg.AddRule(NewConfigRule(30, time.Minute))
		cfg.AddRule(NewConfigRule(100, time.Hour))

		So(cfg.getTickInterval(), ShouldEqual, 100 * time.Millisecond)
	})
}
