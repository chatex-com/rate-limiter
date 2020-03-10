package worker

import (
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/chatex-com/rate-limiter/internal/limiter"
	"github.com/chatex-com/rate-limiter/pkg/config"
	"github.com/chatex-com/rate-limiter/pkg/job"
)

func TestNewWorker(t *testing.T) {
	Convey("Create new worker", t, func() {
		quotas := &limiter.QuotaGroup{}
		requests := make(<-chan job.Request)
		wg := &sync.WaitGroup{}

		worker := NewWorker(quotas, requests, wg)

		So(worker, ShouldNotBeNil)
		So(worker, ShouldHaveSameTypeAs, &Worker{})
		So(worker.quotas, ShouldEqual, quotas)
		So(worker.requests, ShouldEqual, requests)
		So(worker.wg, ShouldEqual, wg)
	})
}

func TestStartStop(t *testing.T) {
	Convey("Start(), Stop(), IsRunning()", t, func() {
		worker := NewWorker(&limiter.QuotaGroup{}, make(chan job.Request), &sync.WaitGroup{})

		So(worker.IsRunning(), ShouldBeFalse)
		worker.Start()
		So(worker.IsRunning(), ShouldBeTrue)
		worker.Stop()
		So(worker.IsRunning(), ShouldBeFalse)
	})

	Convey("Multiple calls Start(), Stop()", t, func() {
		worker := NewWorker(&limiter.QuotaGroup{}, make(chan job.Request), &sync.WaitGroup{})

		So(worker.IsRunning(), ShouldBeFalse)
		worker.Start()
		worker.Start()
		So(worker.IsRunning(), ShouldBeTrue)
		worker.Stop()
		worker.Stop()
		So(worker.IsRunning(), ShouldBeFalse)
	})
}

func TestLoop(t *testing.T) {
	Convey("Success job execution", t, func() {
		quotas, _ := limiter.NewQuotaGroup([]config.Quota{})
		wg := &sync.WaitGroup{}
		requests := make(chan job.Request)
		worker := NewWorker(quotas, requests, wg)
		worker.Start()

		request := job.Request{
			Job: func() (interface{}, error) {
				return 123, nil
			},
			Ch: make(chan job.Response),
		}

		So(worker.stat.Done, ShouldBeZeroValue)

		wg.Add(1)
		requests <- request

		resp := <-request.Ch
		So(resp, ShouldHaveSameTypeAs, job.Response{})
		So(resp.Result, ShouldEqual, 123)
		So(resp.Error, ShouldBeNil)

		wg.Wait()
		So(worker.stat.Done, ShouldEqual, 1)
	})

	Convey("Success job execution after waiting free slot", t, func() {
		quotas, _ := limiter.NewQuotaGroup([]config.Quota{
			*config.NewQuota(1, time.Second),
		})
		wg := &sync.WaitGroup{}
		requests := make(chan job.Request, 2)
		worker := NewWorker(quotas, requests, wg)
		worker.Start()

		request1 := job.Request{
			Job: func() (interface{}, error) {
				time.Sleep(20 * time.Millisecond)
				return nil, nil
			},
			Ch: make(chan job.Response),
		}

		request2 := job.Request{
			Job: func() (interface{}, error) {
				return nil, nil
			},
			Ch: make(chan job.Response),
		}

		So(worker.stat.Error, ShouldBeZeroValue)
		So(worker.stat.Done, ShouldBeZeroValue)

		wg.Add(2)
		requests <- request1
		requests <- request2

		resp1 := <-request1.Ch
		So(resp1, ShouldHaveSameTypeAs, job.Response{})
		So(resp1.Result, ShouldBeNil)
		So(resp1.Error, ShouldBeNil)

		resp2 := <-request2.Ch
		So(resp2, ShouldHaveSameTypeAs, job.Response{})
		So(resp2.Result, ShouldBeNil)
		So(resp2.Error, ShouldBeNil)

		wg.Wait()
		So(worker.stat.Done, ShouldEqual, 2)
	})

	Convey("Error job execution", t, func() {
		quotas, _ := limiter.NewQuotaGroup([]config.Quota{})
		wg := &sync.WaitGroup{}
		requests := make(chan job.Request)
		worker := NewWorker(quotas, requests, wg)
		worker.Start()

		request := job.Request{
			Job: func() (interface{}, error) {
				return nil, job.ErrJobExpired
			},
			Ch: make(chan job.Response),
		}

		So(worker.stat.Error, ShouldBeZeroValue)
		So(worker.stat.Done, ShouldBeZeroValue)

		wg.Add(1)
		requests <- request

		resp := <-request.Ch
		So(resp, ShouldHaveSameTypeAs, job.Response{})
		So(resp.Result, ShouldBeNil)
		So(resp.Error, ShouldBeError)
		So(resp.Error, ShouldEqual, job.ErrJobExpired)

		wg.Wait()
		So(worker.stat.Error, ShouldEqual, 1)
		So(worker.stat.Done, ShouldBeZeroValue)
	})

	Convey("Expired job execution", t, func() {
		quotas, _ := limiter.NewQuotaGroup([]config.Quota{})
		wg := &sync.WaitGroup{}
		requests := make(chan job.Request)
		worker := NewWorker(quotas, requests, wg)
		worker.Start()

		request := job.Request{
			Job: func() (interface{}, error) {

				return nil, nil
			},
			Ch:        make(chan job.Response),
			ExpiredAt: time.Now().Add(-time.Hour),
		}

		So(worker.stat.Error, ShouldBeZeroValue)
		So(worker.stat.Done, ShouldBeZeroValue)

		wg.Add(1)
		requests <- request

		resp := <-request.Ch
		So(resp, ShouldHaveSameTypeAs, job.Response{})
		So(resp.Result, ShouldBeNil)
		So(resp.Error, ShouldBeError)
		So(resp.Error, ShouldEqual, job.ErrJobExpired)

		wg.Wait()
		So(worker.stat.Error, ShouldEqual, 1)
		So(worker.stat.Done, ShouldBeZeroValue)
	})

	Convey("Expired job execution on waiting free slot", t, func() {
		quotas, _ := limiter.NewQuotaGroup([]config.Quota{
			*config.NewQuota(1, time.Second),
		})
		wg := &sync.WaitGroup{}
		requests := make(chan job.Request, 2)
		worker := NewWorker(quotas, requests, wg)
		worker.Start()

		request1 := job.Request{
			Job: func() (interface{}, error) {
				return nil, nil
			},
			Ch: make(chan job.Response),
		}

		request2 := job.Request{
			Job: func() (interface{}, error) {

				return nil, nil
			},
			Ch:        make(chan job.Response),
			ExpiredAt: time.Now().Add(10 * time.Millisecond),
		}

		So(worker.stat.Error, ShouldBeZeroValue)
		So(worker.stat.Done, ShouldBeZeroValue)

		wg.Add(2)
		requests <- request1
		requests <- request2

		resp1 := <-request1.Ch
		So(resp1, ShouldHaveSameTypeAs, job.Response{})
		So(resp1.Result, ShouldBeNil)
		So(resp1.Error, ShouldBeNil)

		resp2 := <-request2.Ch
		So(resp2, ShouldHaveSameTypeAs, job.Response{})
		So(resp2.Result, ShouldBeNil)
		So(resp2.Error, ShouldBeError)
		So(resp2.Error, ShouldEqual, job.ErrJobExpired)

		wg.Wait()
		So(worker.stat.Error, ShouldEqual, 1)
		So(worker.stat.Done, ShouldEqual, 1)
	})
}

func TestReserveFreeSlot(t *testing.T) {
	Convey("Empty queue", t, func() {
		quotas, _ := limiter.NewQuotaGroup([]config.Quota{
			*config.NewQuota(10, time.Second),
		})
		worker := NewWorker(quotas, make(chan job.Request), &sync.WaitGroup{})
		request := job.Request{}

		err := worker.reserveFreeSlot(request)

		So(err, ShouldBeNil)
	})

	Convey("Busy queue with free slots", t, func() {
		quotas, _ := limiter.NewQuotaGroup([]config.Quota{
			*config.NewQuota(10, time.Second),
		})
		worker := NewWorker(quotas, make(chan job.Request), &sync.WaitGroup{})
		request := job.Request{}

		_ = worker.reserveFreeSlot(request)
		err := worker.reserveFreeSlot(request)

		So(err, ShouldBeNil)
	})

	Convey("Busy slot with expired job", t, func() {
		quotas, _ := limiter.NewQuotaGroup([]config.Quota{
			*config.NewQuota(1, 20*time.Millisecond),
		})
		worker := NewWorker(quotas, make(chan job.Request), &sync.WaitGroup{})
		request := job.Request{ExpiredAt: time.Now().Add(- time.Hour)}

		_ = worker.reserveFreeSlot(request)
		err := worker.reserveFreeSlot(request)

		So(err, ShouldBeError)
		So(err, ShouldEqual, job.ErrJobExpired)
	})
}
