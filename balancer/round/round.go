package round

import (
	"github.com/diy-cloud/virtual-gate/balancer"
	"github.com/diy-cloud/virtual-gate/lock"
)

type Round struct {
	candidates []string
	index      int
	l          *lock.Lock
}

func New() balancer.Balancer {
	return &Round{
		candidates: make([]string, 0),
		index:      0,
		l:          new(lock.Lock),
	}
}

func (r *Round) Add(target string) error {
	r.l.Lock()
	defer r.l.Unlock()

	for _, v := range r.candidates {
		if v == target {
			return balancer.ErrorAlreadyExist()
		}
	}

	r.candidates = append(r.candidates, target)
	return nil
}

func (r *Round) Sub(target string) error {
	r.l.Lock()
	defer r.l.Unlock()

	for i, v := range r.candidates {
		if v == target {
			r.candidates = append(r.candidates[:i], r.candidates[i+1:]...)
			return nil
		}
	}

	return balancer.ErrorValueIsNotExist()
}

func (r *Round) Get(_ string) (string, error) {
	r.l.Lock()
	defer r.l.Unlock()

	if len(r.candidates) == 0 {
		return "", balancer.ErrorAnythingNotExist()
	}

	if r.index >= len(r.candidates) {
		r.index = 0
	}
	target := r.candidates[r.index]
	r.index = r.index + 1
	return target, nil
}

func (r *Round) Restore(_ string) error {
	return nil
}
