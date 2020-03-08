package config

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewQuota(t *testing.T) {
	Convey("Empty settings (defaults)", t, func() {
		rule := NewQuota(10, time.Second)

		So(rule.Capacity, ShouldEqual, 10)
		So(rule.Interval, ShouldEqual, time.Second)
	})
}
