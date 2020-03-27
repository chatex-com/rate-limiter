# Rate Limiter

<a href="https://opensource.org/licenses/Apache-2.0" rel="nofollow"><img src="https://img.shields.io/badge/license-Apache%202-blue" alt="License" style="max-width:100%;"></a>
![unit-tests](https://github.com/chatex-com/rate-limiter/workflows/unit-tests/badge.svg)

## Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/chatex-com/rate-limiter"
	"github.com/chatex-com/rate-limiter/pkg/config"
	"github.com/chatex-com/rate-limiter/pkg/job"
)

func main() {
	cfg := config.NewConfigWithQuotas([]*config.Quota{
		config.NewQuota(10, time.Second),
		config.NewQuota(20, time.Minute),
	})
	cfg.Concurrency = 10

	rateLimiter, _ := limiter.NewRateLimiter(cfg)
	rateLimiter.Start()

	var ch <-chan job.Response
	var response job.Response

	// Execute job when it will be allowed by quota
	ch = rateLimiter.Execute(func() (interface{}, error) {
		return nil, nil
	})

	response = <-ch
	fmt.Println(response.Result) // nil
	fmt.Println(response.Error) // nil

	// Execute job when it will be allowed by quota with timeout
	ch = rateLimiter.ExecuteWithTimout(func() (interface{}, error) {
		return nil, nil
	}, time.Minute)

	response = <-ch
	fmt.Println(response.Result) // nil
	fmt.Println(response.Error) // nil

	rateLimiter.AwaitAll()
	rateLimiter.Stop()
}
```
