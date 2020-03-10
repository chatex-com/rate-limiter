package limiter

import (
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/chatex-com/rate-limiter/internal/limiter"
	"github.com/chatex-com/rate-limiter/pkg/config"
	"github.com/chatex-com/rate-limiter/pkg/job"
)

func TestNewRateLimiter(t *testing.T) {
	Convey("success", t, func() {
		cfg := config.NewConfig()
		cfg.Concurrency = 5
		l, err := NewRateLimiter(cfg)

		So(err, ShouldBeNil)
		So(l, ShouldHaveSameTypeAs, &RateLimiter{})
		So(l.workers, ShouldHaveLength, 5)
	})

	Convey("wrong configuration", t, func() {
		cfg := config.NewConfig()
		cfg.AddQuota(config.NewQuota(0, 0))
		l, err := NewRateLimiter(cfg)

		So(err, ShouldBeError)
		So(err, ShouldBeIn, []error{limiter.ErrZeroRuleCount, limiter.ErrZeroRuleInterval})
		So(l, ShouldBeNil)
	})
}

func TestRateLimiter_Execute(t *testing.T) {
	Convey("success job execution", t, func() {
		cfg := config.NewConfig()
		cfg.Concurrency = 1
		l, _ := NewRateLimiter(cfg)
		l.Start()

		ch := l.Execute(func() (interface{}, error) {
			return "foo", nil
		})

		resp := <-ch
		So(resp.Error, ShouldBeNil)
		So(resp.Result, ShouldEqual, "foo")
	})

	Convey("job execution with timeout", t, func() {
		cfg := config.NewConfig()
		cfg.Concurrency = 1
		l, _ := NewRateLimiter(cfg)
		l.Start()

		ch := l.ExecuteWithTimout(func() (interface{}, error) {
			return "foo", nil
		}, time.Second)

		resp := <-ch
		So(resp.Error, ShouldBeNil)
		So(resp.Result, ShouldEqual, "foo")
	})

	Convey("error job execution", t, func() {
		cfg := config.NewConfig()
		cfg.Concurrency = 1
		l, _ := NewRateLimiter(cfg)
		l.Start()

		ch := l.Execute(func() (interface{}, error) {
			return nil, job.ErrJobExpired
		})

		resp := <-ch
		So(resp.Error, ShouldBeError)
		So(resp.Error, ShouldEqual, job.ErrJobExpired)
		So(resp.Result, ShouldBeNil)
	})
}

func TestStartStop(t *testing.T) {
	Convey("Start(), Stop() all workers", t, func() {
		cfg := config.NewConfig()
		cfg.Concurrency = 2
		l, _ := NewRateLimiter(cfg)

		So(l.isRunning, ShouldBeFalse)
		So(l.workers[0].IsRunning(), ShouldBeFalse)
		So(l.workers[1].IsRunning(), ShouldBeFalse)

		l.Start()

		So(l.isRunning, ShouldBeTrue)
		So(l.workers[0].IsRunning(), ShouldBeTrue)
		So(l.workers[1].IsRunning(), ShouldBeTrue)

		l.Start()

		So(l.isRunning, ShouldBeTrue)
		So(l.workers[0].IsRunning(), ShouldBeTrue)
		So(l.workers[1].IsRunning(), ShouldBeTrue)

		l.Stop()

		So(l.isRunning, ShouldBeFalse)
		So(l.workers[0].IsRunning(), ShouldBeFalse)
		So(l.workers[1].IsRunning(), ShouldBeFalse)

		l.Stop()

		So(l.isRunning, ShouldBeFalse)
		So(l.workers[0].IsRunning(), ShouldBeFalse)
		So(l.workers[1].IsRunning(), ShouldBeFalse)
	})
}

func TestAwaitAll(t *testing.T) {
	Convey("Waiting till all jobs are executed", t, func() {
		cfg := config.NewConfig()
		cfg.Concurrency = 1
		l, _ := NewRateLimiter(cfg)
		l.Start()

		ch := l.Execute(func() (interface{}, error) {
			time.Sleep(50 * time.Millisecond)

			return true, nil
		})

		var wg sync.WaitGroup
		wg.Add(1)
		var res job.Response
		go func() {
			res = <-ch
			wg.Done()
		}()

		l.AwaitAll()
		l.Stop()
		So(l.isRunning, ShouldBeFalse)

		wg.Wait()
		So(res.Result, ShouldBeTrue)
	})
}
