package limiter

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"rate_limiter/pkg/config"
)

func TestNewRunnerRule(t *testing.T) {
	Convey("Error on rule.Capacity is zero", t, func() {
		rule, err := NewQuota(*config.NewQuota(0, time.Second))

		So(rule, ShouldBeNil)
		So(err, ShouldBeError)
		So(err, ShouldEqual, ErrZeroRuleCount)
	})

	Convey("Error on rule.Interval is zero", t, func() {
		rule, err := NewQuota(*config.NewQuota(10, 0))

		So(rule, ShouldBeNil)
		So(err, ShouldBeError)
		So(err, ShouldEqual, ErrZeroRulePeriod)
	})

	Convey("Success creation", t, func() {
		cfg := *config.NewQuota(10, time.Second)
		quota, err := NewQuota(cfg)

		So(err, ShouldBeNil)
		So(quota.cfg, ShouldHaveSameTypeAs, cfg)
		So(quota.cfg.Capacity, ShouldEqual, cfg.Capacity)
		So(quota.cfg.Interval, ShouldEqual, cfg.Interval)
	})
}

func TestAddTime(t *testing.T) {
	Convey("Add time", t, func() {
		rule, _ := NewQuota(*config.NewQuota(1, 10 * time.Millisecond))
		rule.Add(time.Now())

		So(rule.times, ShouldHaveLength, 1)

		time.Sleep(20 * time.Millisecond)

		So(rule.times, ShouldBeEmpty)
	})
}

func TestFreeSlots(t *testing.T) {
	Convey("default", t, func() {
		rule, _ := NewQuota(*config.NewQuota(2, 10 * time.Millisecond))

		So(rule.freeSlots(), ShouldEqual, 2)

		rule.Add(time.Now())

		So(rule.freeSlots(), ShouldEqual, 1)

		time.Sleep(20 * time.Millisecond)

		So(rule.freeSlots(), ShouldEqual, 2)

		rule.Add(time.Now())
		rule.Add(time.Now())

		So(rule.freeSlots(), ShouldEqual, 0)

		time.Sleep(20 * time.Millisecond)

		So(rule.freeSlots(), ShouldEqual, 2)
	})
}

func TestGetFreeSlot(t *testing.T) {
	Convey("empty queue", t, func() {
		rule, _ := NewQuota(*config.NewQuota(1, time.Second))

		wait, exist := rule.GetFreeSlot()

		So(exist, ShouldBeTrue)
		So(wait, ShouldBeZeroValue)
	})

	Convey("busy queue", t, func() {
		rule, _ := NewQuota(*config.NewQuota(1, time.Second))

		rule.Add(time.Now())

		wait, exist := rule.GetFreeSlot()

		So(exist, ShouldBeFalse)
		So(wait, ShouldAlmostEqual, time.Second, time.Millisecond)
	})

	Convey("busy queue -> free queue", t, func() {
		rule, _ := NewQuota(*config.NewQuota(1, 10 * time.Millisecond))

		rule.Add(time.Now())

		wait, exist := rule.GetFreeSlot()
		So(exist, ShouldBeFalse)
		So(wait, ShouldAlmostEqual, 10 * time.Millisecond, time.Millisecond)

		time.Sleep(50 * wait)

		wait, exist = rule.GetFreeSlot()
		So(exist, ShouldBeTrue)
		So(wait, ShouldBeZeroValue)
	})
}
