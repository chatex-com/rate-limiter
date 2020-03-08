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
)

func main() {
	cfg := config.NewConfigWithQuotas([]*config.Quota{
		config.NewQuota(10, time.Second),
		config.NewQuota(20, time.Minute),
	})
	cfg.Concurrency = 10

	limiter, _ := rate_limiter.NewRateLimiter(*cfg)
	limiter.Start()

	ch := limiter.Execute(func() (interface{}, error) {
		return nil, nil
	})

	r := <-ch
	fmt.Println(r.Result) // nil
	fmt.Println(r.Error) // nil

	limiter.AwaitAll()
	limiter.Stop()
}
```
