package simple_breaker

import (
	"github.com/diy-cloud/virtual-gate/circuit_breaker"
	"github.com/diy-cloud/virtual-gate/lock"
)

type Breaker struct {
	cache map[string]struct{}
	l     *lock.Lock
}

func New() circuit_breaker.CurciutBreaker {
	return &Breaker{
		cache: make(map[string]struct{}),
		l:     new(lock.Lock),
	}
}

func (b *Breaker) BreakDown(target string) error {
	b.l.Lock()
	defer b.l.Unlock()
	b.cache[target] = struct{}{}
	return nil
}

func (b *Breaker) Restore(target string) error {
	b.l.Lock()
	defer b.l.Unlock()
	delete(b.cache, target)
	return nil
}

func (b *Breaker) IsBrokeDown(target string) bool {
	b.l.Lock()
	defer b.l.Unlock()
	_, ok := b.cache[target]
	return ok
}
