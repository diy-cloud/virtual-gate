package breaker

import "errors"

type CurciutBreaker interface {
	BreakDown(target string) error
	Restore(target string) error
	IsBrokeDown(target string) bool
}

var notFoundErr = errors.New("not found")

func ErrorNotFound() error {
	return notFoundErr
}
