package job_runner

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsExpired(t *testing.T) {
	Convey("Test never expired option", t, func() {
		r := JobRequest{}

		So(r.isExpired(), ShouldBeFalse)
	})

	Convey("Test expiration in 10 ms", t, func() {
		r := JobRequest{
			expiredAt: time.Now().Add(10 * time.Millisecond),
		}

		time.Sleep(20 * time.Millisecond)

		So(r.isExpired(), ShouldBeTrue)
	})

	Convey("Test non expired item", t, func() {
		r := JobRequest{
			expiredAt: time.Now().Add(time.Hour),
		}

		So(r.isExpired(), ShouldBeFalse)
	})
}
