package acl

import (
	"time"

	"github.com/diy-cloud/virtual-gate/limiter"
	"github.com/diy-cloud/virtual-gate/limiter/slide_count"
	"github.com/diy-cloud/virtual-gate/lock"
)

type AccessControlList struct {
	lock       *lock.Lock
	list       map[string]limiter.Limiter
	maxConnPer float64
	unit       time.Duration
}

func New(maxConnPer float64, unit time.Duration) limiter.Limiter {
	return &AccessControlList{
		lock:       new(lock.Lock),
		list:       nil,
		maxConnPer: maxConnPer,
		unit:       unit,
	}
}

func (acl *AccessControlList) TryTake(key []byte) (bool, int) {
	acl.lock.Lock()
	defer acl.lock.Unlock()

	if acl.list == nil {
		acl.list = make(map[string]limiter.Limiter)
	}

	slide, ok := acl.list[string(key)]
	if !ok {
		slide = slide_count.New(acl.maxConnPer, acl.unit)
		acl.list[string(key)] = slide
	}

	return slide.TryTake(nil)
}
