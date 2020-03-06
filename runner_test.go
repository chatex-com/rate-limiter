package job_runner

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewRunner(t *testing.T) {
	Convey("empty configuration", t, func() {
		cfg := NewConfig()
		runner, err := NewRunner(cfg)

		So(err, ShouldBeNil)
		So(runner, ShouldNotBeNil)
		So(cap(runner.requests), ShouldEqual, defaultConcurrencyLimit)
		So(runner.rules, ShouldBeEmpty)
	})

	Convey("with rules", t, func() {
		cfg := NewConfigWithRules([]*ConfigRule{
			NewConfigRule(10, time.Second),
			NewConfigRule(30, time.Minute),
			NewConfigRule(100, time.Hour),
		})
		runner, _ := NewRunner(cfg)

		So(runner.rules, ShouldHaveLength, 3)
	})

	Convey("error in configuration", t, func() {
		cfg := NewConfigWithRules([]*ConfigRule{
			NewConfigRule(0, 0),
		})
		runner, err := NewRunner(cfg)

		So(err, ShouldBeError)
		So(err, ShouldBeIn, []error{ErrZeroRuleCount, ErrZeroRulePeriod})
		So(runner, ShouldBeNil)
	})
}

func TestHasConcurrencySlot(t *testing.T) {
	Convey("has concurrency slot on empty queue", t, func() {
		cfg := NewConfigWithRules([]*ConfigRule{
			NewConfigRule(1, time.Second),
		})

		runner, _ := NewRunner(cfg)

		So(runner.hasConcurrentSlot(), ShouldBeTrue)
	})

	Convey("has concurrency slot on busy queue", t, func() {
		runner, _ := NewRunner(NewConfigWithRules([]*ConfigRule{
			NewConfigRule(10, time.Second),
		}))

		_ = runner.Execute(func() (interface{}, error) {
			time.Sleep(10 * time.Millisecond)

			return nil, nil
		})

		time.Sleep(minimalTickInterval + time.Millisecond) // Wait for starting job
		So(runner.hasConcurrentSlot(), ShouldBeTrue)
	})

	Convey("has concurrency slot on full queue", t, func() {
		cfg := NewConfigWithRules([]*ConfigRule{
			NewConfigRule(10, time.Second),
		})
		cfg.ConcurrencyLimit = 1
		runner, _ := NewRunner(cfg)

		_ = runner.Execute(func() (interface{}, error) {
			time.Sleep(10 * time.Millisecond)

			return nil, nil
		})

		time.Sleep(minimalTickInterval + time.Millisecond) // Wait for starting job
		So(runner.hasConcurrentSlot(), ShouldBeFalse)

		time.Sleep(20 * time.Millisecond)
		So(runner.hasConcurrentSlot(), ShouldBeTrue)
	})
}

func TestExecuteRequest(t *testing.T) {
	Convey("check stats on request execution", t, func() {
		runner, _ := NewRunner(NewConfig())

		_ = runner.Execute(func() (interface{}, error) {
			time.Sleep(20 * time.Millisecond)

			return nil, nil
		})

		time.Sleep(minimalTickInterval + time.Millisecond) // Wait for starting job
		stat := runner.Stat()

		So(stat.expired, ShouldBeZeroValue)
		So(stat.queue, ShouldBeZeroValue)
		So(stat.inProgress, ShouldEqual, 1)
		So(stat.done, ShouldBeZeroValue)

		time.Sleep(30 * time.Millisecond) // Wait for starting job
		stat = runner.Stat()

		So(stat.expired, ShouldBeZeroValue)
		So(stat.queue, ShouldBeZeroValue)
		So(stat.inProgress, ShouldBeZeroValue)
		So(stat.done, ShouldEqual, 1)
	})

	Convey("check result of execution", t, func() {
		runner, _ := NewRunner(NewConfig())

		Convey("scalar value", func() {
			chUint := runner.Execute(func() (interface{}, error) {
				return uint(123), nil
			})
			chString := runner.Execute(func() (interface{}, error) {
				return "foo bar", nil
			})

			var r JobResponse

			r = <-chUint
			So(r.Result, ShouldEqual, uint(123))

			r = <-chString
			So(r.Result, ShouldEqual, "foo bar")
		})

		Convey("struct", func() {
			type FooType struct {
				foo string
				bar int
			}

			expected := FooType{foo: "baz", bar: 12}

			ch := runner.Execute(func() (interface{}, error) {
				return expected, nil
			})

			r := <-ch
			So(r.Result, ShouldHaveSameTypeAs, expected)

			v := r.Result.(FooType)
			So(v.foo, ShouldEqual, expected.foo)
			So(v.bar, ShouldEqual, expected.bar)
		})
	})

	Convey("wait execution when concurrency limit is reached", t, func() {
		cfg := NewConfig()
		cfg.ConcurrencyLimit = 1
		runner, _ := NewRunner(cfg)

		_ = runner.Execute(func() (interface{}, error) {
			time.Sleep(20 * time.Millisecond)
			return nil, nil
		})

		_ = runner.Execute(func() (interface{}, error) {
			time.Sleep(20 * time.Millisecond)
			return nil, nil
		})

		time.Sleep(100 * time.Millisecond)

		stat := runner.Stat()
		So(stat.done, ShouldEqual, 2)
	})

	Convey("check job expiration", t, func() {
		cfg := NewConfig()
		cfg.ConcurrencyLimit = 1
		runner, _ := NewRunner(cfg)

		_ = runner.Execute(func() (interface{}, error) {
			time.Sleep(100 * time.Millisecond)

			return nil, nil
		})

		// FIXME [ ]: In immediately strategy we have to execute job as soon as possible
		//            We shouldn't wait for minimalTickInterval

		time.Sleep(minimalTickInterval + time.Millisecond) // Wait for starting job
		So(runner.hasConcurrentSlot(), ShouldBeFalse)

		ch := runner.ExecuteWithExpiration(func() (interface{}, error) {
			return nil, nil
		}, 20 * time.Millisecond)

		res := <-ch

		So(res.Error, ShouldBeError)
		So(res.Error, ShouldEqual, ErrJobExpired)

		time.Sleep(time.Millisecond)

		stat := runner.Stat()
		So(stat.expired, ShouldEqual, 1)
		So(stat.queue, ShouldBeZeroValue)
		So(stat.inProgress, ShouldBeZeroValue)
		So(stat.done, ShouldEqual, 1)
	})
}
