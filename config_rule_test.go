package job_runner

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewConfigRule(t *testing.T) {
	Convey("Empty settings (defaults)", t, func() {
		rule := NewConfigRule(10, time.Second)

		So(rule.Count, ShouldEqual, 10)
		So(rule.Period, ShouldEqual, time.Second)
	})
}

func TestGetInterval(t *testing.T) {
	Convey("Usual case", t, func() {
		rule := NewConfigRule(10, time.Second)

		So(rule.getInterval(), ShouldEqual, 100 * time.Millisecond)
	})

	Convey("If tick less than minimal tick interval", t, func() {
		rule := NewConfigRule(10, minimalTickInterval)

		So(rule.getInterval(), ShouldEqual, minimalTickInterval)
	})
}
