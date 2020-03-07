package job

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsExpired(t *testing.T) {
	Convey("Test never expired option", t, func() {
		r := Request{}

		So(r.IsExpired(), ShouldBeFalse)
	})

	Convey("Test expiration in 10 ms", t, func() {
		r := Request{
			ExpiredAt: time.Now().Add(10 * time.Millisecond),
		}

		time.Sleep(20 * time.Millisecond)

		So(r.IsExpired(), ShouldBeTrue)
	})

	Convey("Test non expired item", t, func() {
		r := Request{
			ExpiredAt: time.Now().Add(time.Hour),
		}

		So(r.IsExpired(), ShouldBeFalse)
	})
}
