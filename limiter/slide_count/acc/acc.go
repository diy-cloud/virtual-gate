package acc

import (
	"time"

	"github.com/diy-cloud/virtual-gate/limiter"
	"github.com/diy-cloud/virtual-gate/limiter/slide_count"
	"github.com/diy-cloud/virtual-gate/lock"
)

type AccessControlCount struct {
	lock       *lock.Lock
	list       map[string]limiter.Limiter
	maxConnPer float64
	unit       time.Duration
}

func New(maxConnPer float64, unit time.Duration) limiter.Limiter {
	return &AccessControlCount{
		lock:       new(lock.Lock),
		list:       nil,
		maxConnPer: maxConnPer,
		unit:       unit,
	}
}

func (acc *AccessControlCount) TryTake(key []byte) (bool, int) {
	acc.lock.Lock()
	defer acc.lock.Unlock()

	if acc.list == nil {
		acc.list = make(map[string]limiter.Limiter)
	}

	slide, ok := acc.list[string(key)]
	if !ok {
		slide = slide_count.New(acc.maxConnPer, acc.unit)
		acc.list[string(key)] = slide
	}

	return slide.TryTake(nil)
}
