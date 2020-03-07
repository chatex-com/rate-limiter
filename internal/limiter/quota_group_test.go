package limiter

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"rate_limiter/pkg/config"
)

func TestNewQuotaGroup(t *testing.T) {
	Convey("Error on creation quotas group", t, func() {
		cfg := config.NewConfig()
		cfg.AddQuota(config.NewQuota(0, 0))
		group, err := NewQuotaGroup(cfg.GetQuotas())

		So(err, ShouldBeError)
		So(err, ShouldBeIn, []error{ErrZeroRuleInterval, ErrZeroRuleCount})
		So(group, ShouldBeNil)
	})

	Convey("Success creation", t, func() {
		cfg := config.NewConfigWithQuotas([]*config.Quota{
			config.NewQuota(10, time.Second),
			config.NewQuota(60, time.Minute),
		})
		group, err := NewQuotaGroup(cfg.GetQuotas())

		So(err, ShouldBeNil)
		So(group, ShouldHaveSameTypeAs, &QuotaGroup{})
		So(group.quotas, ShouldHaveLength, 2)
	})
}

func TestCreateList(t *testing.T) {
	Convey("Create empty list", t, func() {
		var quotas []config.Quota
		list, err := createList(quotas)

		So(err, ShouldBeNil)
		So(list, ShouldHaveLength, 0)
	})

	Convey("Create quotas list from config", t, func() {
		q1 := config.Quota{Capacity: 10, Interval: time.Second}
		q2 := config.Quota{Capacity: 60, Interval: time.Minute}
		quotas := []config.Quota{
			q1,
			q2,
		}

		list, err := createList(quotas)

		So(err, ShouldBeNil)
		So(list, ShouldHaveLength, 2)
		So(list[0].cfg.Interval, ShouldEqual, q1.Interval)
		So(list[0].cfg.Capacity, ShouldEqual, q1.Capacity)
		So(list[1].cfg.Interval, ShouldEqual, q2.Interval)
		So(list[1].cfg.Capacity, ShouldEqual, q2.Capacity)
	})
}

func TestReverse(t *testing.T) {
	Convey("", t, func() {
		quotas := []config.Quota{
			config.Quota{Capacity: 10, Interval: time.Second},
			config.Quota{Capacity: 60, Interval: time.Minute},
		}

		now := time.Now()
		check := time.Since(now) + time.Millisecond
		group, _ := NewQuotaGroup(quotas)
		group.reserve()

		time.Sleep(time.Millisecond) // check that all goroutines were started

		So(time.Since(group.quotas[0].times[0]), ShouldAlmostEqual, check, time.Millisecond)
		So(time.Since(group.quotas[1].times[0]), ShouldAlmostEqual, check, time.Millisecond)
	})
}

func TestReserveFreeSlot(t *testing.T) {
	Convey("Empty quotas", t, func() {
		group, _ := NewQuotaGroup([]config.Quota{})
		free, wait := group.ReserveFreeSlot()

		So(free, ShouldBeTrue)
		So(wait, ShouldEqual, 0)
	})
}
