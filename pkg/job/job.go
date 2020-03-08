package job

import (
	"errors"
)

var ErrJobExpired = errors.New("job was expired")

type Job func() (interface{}, error)

type Response struct {
	Result interface{}
	Error  error
}
