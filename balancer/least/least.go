package least

import (
	"math"

	"github.com/diy-cloud/virtual-gate/balancer"
	"github.com/diy-cloud/virtual-gate/lock"
)

type Least struct {
	candidates map[string]int64
	l          *lock.Lock
}

func New() balancer.Balancer {
	return &Least{
		candidates: make(map[string]int64),
		l:          new(lock.Lock),
	}
}

func (l *Least) Add(target string) error {
	l.l.Lock()
	defer l.l.Unlock()

	if _, ok := l.candidates[target]; ok {
		return balancer.ErrorAlreadyExist()
	}

	l.candidates[target] = 0
	return nil
}

func (l *Least) Sub(target string) error {
	l.l.Lock()
	defer l.l.Unlock()

	if _, ok := l.candidates[target]; !ok {
		return balancer.ErrorValueIsNotExist()
	}

	delete(l.candidates, target)
	return nil
}

func (l *Least) Get(_ string) (string, error) {
	l.l.Lock()
	defer l.l.Unlock()

	min := int64(math.MaxInt64)
	target := ""
	for k, v := range l.candidates {
		if v < min {
			target = k
			min = v
		}
	}

	if target == "" {
		return "", balancer.ErrorAnythingNotExist()
	}

	l.candidates[target]++
	return target, nil
}

func (l *Least) Restore(target string) error {
	l.l.Lock()
	defer l.l.Unlock()

	if _, ok := l.candidates[target]; !ok {
		return balancer.ErrorValueIsNotExist()
	}

	l.candidates[target]--
	return nil
}
