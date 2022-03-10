package count_breaker

import (
	"math/rand"

	"github.com/diy-cloud/virtual-gate/breaker"
	"github.com/diy-cloud/virtual-gate/lock"
)

type CountBreaker struct {
	cache       map[string]int
	maxCount    int
	minimumRate float64
	l           *lock.Lock
}

func New(maxCount int, minimumRate float64) breaker.Breaker {
	return &CountBreaker{
		cache:       make(map[string]int),
		maxCount:    maxCount,
		minimumRate: minimumRate,
		l:           new(lock.Lock),
	}
}

func (c *CountBreaker) BreakDown(target string) error {
	c.l.Lock()
	defer c.l.Unlock()
	if c.cache[target] < c.maxCount {
		c.cache[target] += 1
	}
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
	fTarget := float64(c.cache[target])
	if fTarget >= fMaxCount {
		fTarget = fMaxCount
	}
	maxRate := ((fMaxCount-fTarget)/fMaxCount)*(100-c.minimumRate) + c.minimumRate
	rnd := rand.Float64() * 100
	return rnd >= maxRate
}
