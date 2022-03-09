package hashed

import (
	"hash"

	"github.com/diy-cloud/virtual-gate/balancer"
	"github.com/diy-cloud/virtual-gate/lock"
)

type Hashed struct {
	candidates []string
	hasher     hash.Hash64
	index      int
	l          *lock.Lock
}

func New(hasher hash.Hash64) balancer.Balancer {
	return &Hashed{
		candidates: make([]string, 0, 8),
		hasher:     hasher,
		l:          new(lock.Lock),
	}
}

func (h *Hashed) Add(target string) error {
	h.l.Lock()
	defer h.l.Unlock()

	firstTombstoneIndex := -1
	for i, candidate := range h.candidates {
		if candidate == target {
			return balancer.ErrorAlreadyExist()
		}
		if candidate == "" && firstTombstoneIndex == -1 {
			firstTombstoneIndex = i
		}
	}

	if firstTombstoneIndex != -1 {
		h.candidates[firstTombstoneIndex] = target
		return nil
	}

	h.candidates = append(h.candidates, target)

	return nil
}

func (h *Hashed) Sub(target string) error {
	h.l.Lock()
	defer h.l.Unlock()

	for i, candidate := range h.candidates {
		if candidate == target {
			h.candidates[i] = ""
			return nil
		}
	}

	return balancer.ErrorValueIsNotExist()
}

func (h *Hashed) Get(id string) (string, error) {
	h.l.Lock()
	defer h.l.Unlock()

	h.hasher.Reset()
	h.hasher.Write([]byte(id))
	hashedIndex := int(h.hasher.Sum64() % uint64(len(h.candidates)))
	count := 0
	for {
		count++
		if count >= len(h.candidates) {
			return "", balancer.ErrorNoAvaliableTarget()
		}
		if hashedIndex >= len(h.candidates) {
			hashedIndex = 0
		}
		candidate := h.candidates[hashedIndex]
		if candidate != "" {
			return candidate, nil
		}
		hashedIndex = hashedIndex + 1
	}
}

func (h *Hashed) Restore(_ string) error {
	return nil
}
