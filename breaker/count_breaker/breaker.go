package count_breaker

import (
	"math/rand"

	"github.com/diy-cloud/virtual-gate/breaker"
	"github.com/diy-cloud/virtual-gate/lock"
)

type CountBreaker struct {
	cache    map[string]int
	maxCount int
	l        *lock.Lock
}

func New(maxCount int) breaker.CurciutBreaker {
	return &CountBreaker{
		cache:    make(map[string]int),
		maxCount: maxCount,
		l:        new(lock.Lock),
	}
}

func (c *CountBreaker) BreakDown(target string) error {
	c.l.Lock()
	defer c.l.Unlock()
	c.cache[target] += 1
	return nil
}

func (c *CountBreaker) Restore(target string) error {
	c.l.Lock()
	defer c.l.Unlock()
	if c.cache[target] > 0 {
		c.cache[target] -= 1
	}
	return nil
}

func (c *CountBreaker) IsBrokeDown(target string) bool {
	c.l.Lock()
	defer c.l.Unlock()
	fMaxCount := float64(c.maxCount)
	maxRate := ((fMaxCount-float64(c.cache[target]))/fMaxCount)*90 + 10
	rnd := rand.Float64() * 100
	return rnd >= maxRate
}
