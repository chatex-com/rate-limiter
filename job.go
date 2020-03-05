package job_runner

import (
	"time"
)

type Job func() (interface{}, error)

type JobResponse struct {
	Result interface{}
	Error  error
}

type JobRequest struct {
	job       Job
	ch        chan JobResponse
	expiredAt time.Time
}

func (r JobRequest) isExpired() bool {
	if r.expiredAt.IsZero() {
		return false
	}

	return r.expiredAt.Before(time.Now())
}
