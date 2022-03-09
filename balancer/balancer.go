package balancer

import "errors"

type Balancer interface {
	Add(string) error
	Sub(string) error
	Get(string) (string, error)
	Restore(string) error
}

var alreadyExistErr = errors.New("already exist")

func ErrorAlreadyExist() error {
	return alreadyExistErr
}

var anythingNotExistErr = errors.New("anything not exist")

func ErrorAnythingNotExist() error {
	return anythingNotExistErr
}

var valueIsNotExistErr = errors.New("value is not exist")

func ErrorValueIsNotExist() error {
	return valueIsNotExistErr
}

var noAvaliableTargetErr = errors.New("no avaliable target")

func ErrorNoAvaliableTarget() error {
	return noAvaliableTargetErr
}
