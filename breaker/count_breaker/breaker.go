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
	rnd := rand.Int63n(int64(c.maxCount * 120 / 100))
	if rnd == 1 {
		rnd++
	}
	if c.cache[target] > c.maxCount {
		c.cache[target] = c.maxCount
	}
	return rnd < int64(c.cache[target])
}
