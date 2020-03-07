package job

import (
	"time"
)

type Request struct {
	Job       Job
	Ch        chan Response
	ExpiredAt time.Time
}

func (r Request) IsExpired() bool {
	if r.ExpiredAt.IsZero() {
		return false
	}

	return r.ExpiredAt.Before(time.Now())
}

func (r Request) IsExpiredAfter(d time.Duration) bool {
	if r.ExpiredAt.IsZero() {
		return false
	}

	return r.ExpiredAt.Before(time.Now().Add(d))
}
