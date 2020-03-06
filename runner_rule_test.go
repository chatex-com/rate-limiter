package job_runner

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewRunnerRule(t *testing.T) {
	Convey("Error on rule.Count is zero", t, func() {
		rule, err := newRunnerRule(NewConfigRule(0, time.Second))

		So(rule, ShouldBeNil)
		So(err, ShouldBeError)
		So(err, ShouldEqual, ErrZeroRuleCount)
	})

	Convey("Error on rule.Period is zero", t, func() {
		rule, err := newRunnerRule(NewConfigRule(10, 0))

		So(rule, ShouldBeNil)
		So(err, ShouldBeError)
		So(err, ShouldEqual, ErrZeroRulePeriod)
	})

	Convey("Success creation", t, func() {
		r := NewConfigRule(10, time.Second)
		rule, err := newRunnerRule(r)

		So(err, ShouldBeNil)
		So(rule.cfg, ShouldEqual, r)
	})
}

func TestAddTime(t *testing.T) {
	Convey("Add time", t, func() {
		rule, _ := newRunnerRule(NewConfigRule(1, 10 * time.Millisecond))
		rule.add(time.Now())

		So(rule.times, ShouldHaveLength, 1)

		time.Sleep(20 * time.Millisecond)

		So(rule.times, ShouldBeEmpty)
	})
}

func TestFreeSlots(t *testing.T) {
	Convey("default", t, func() {
		rule, _ := newRunnerRule(NewConfigRule(2, 10 * time.Millisecond))

		So(rule.freeSlots(), ShouldEqual, 2)

		rule.add(time.Now())

		So(rule.freeSlots(), ShouldEqual, 1)

		time.Sleep(20 * time.Millisecond)

		So(rule.freeSlots(), ShouldEqual, 2)

		rule.add(time.Now())
		rule.add(time.Now())

		So(rule.freeSlots(), ShouldEqual, 0)

		time.Sleep(20 * time.Millisecond)

		So(rule.freeSlots(), ShouldEqual, 2)
	})
}

func TestGetFreeSlot(t *testing.T) {
	Convey("empty queue", t, func() {
		rule, _ := newRunnerRule(NewConfigRule(1, time.Second))

		wait, exist := rule.getFreeSlot()

		So(exist, ShouldBeTrue)
		So(wait, ShouldBeZeroValue)
	})

	Convey("busy queue", t, func() {
		rule, _ := newRunnerRule(NewConfigRule(1, time.Second))

		rule.add(time.Now())

		wait, exist := rule.getFreeSlot()

		So(exist, ShouldBeFalse)
		So(wait, ShouldAlmostEqual, time.Second, time.Millisecond)
	})

	Convey("busy queue -> free queue", t, func() {
		rule, _ := newRunnerRule(NewConfigRule(1, 10 * time.Millisecond))

		rule.add(time.Now())

		wait, exist := rule.getFreeSlot()
		So(exist, ShouldBeFalse)
		So(wait, ShouldAlmostEqual, 10 * time.Millisecond, time.Millisecond)

		time.Sleep(50 * time.Millisecond + wait)

		wait, exist = rule.getFreeSlot()
		So(exist, ShouldBeTrue)
		So(wait, ShouldBeZeroValue)
	})
}
