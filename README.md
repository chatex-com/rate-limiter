# Rate Limiter

<a href="https://opensource.org/licenses/Apache-2.0" rel="nofollow"><img src="https://img.shields.io/badge/license-Apache%202-blue" alt="License" style="max-width:100%;"></a>
![unit-tests](https://github.com/chatex-com/rate-limiter/workflows/unit-tests/badge.svg)

## Example

```go
package main

import (
	"fmt"
	"time"

	"rate_limiter"
	"rate_limiter/pkg/config"
	"rate_limiter/pkg/job"
)

func main() {
	cfg := config.NewConfigWithQuotas([]*config.Quota{
		config.NewQuota(10, time.Second),
		config.NewQuota(20, time.Minute),
	})
	cfg.Concurrency = 10

	limiter, _ := rate_limiter.NewRateLimiter(*cfg)
	limiter.Start()

	var ch <-chan job.Response
	var response job.Response
	// Execute job when it will be allowed by quota
	ch = limiter.Execute(func() (interface{}, error) {
		return nil, nil
	})

	response = <-ch
	fmt.Println(response.Result) // nil
	fmt.Println(response.Error) // nil

	// Execute job when it will be allowed by quota with timeout
	ch = limiter.ExecuteWithTimout(func() (interface{}, error) {
		return nil, nil
	}, time.Minute)

	response = <-ch
	fmt.Println(response.Result) // nil
	fmt.Println(response.Error) // nil

	limiter.AwaitAll()
	limiter.Stop()
}
```
