package job_runner

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConstantToString(t *testing.T) {
	Convey("Test constant to string conversion", t, func() {
		So(fmt.Sprintf("%s", StrategyImmediately), ShouldEqual, "Immediately")
		So(fmt.Sprintf("%s", StrategyEvenly), ShouldEqual, "Evenly")
	})
}
