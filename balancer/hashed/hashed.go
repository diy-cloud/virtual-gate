package hashed

import (
	"encoding/binary"

	"github.com/diy-cloud/virtual-gate/balancer"
	"github.com/diy-cloud/virtual-gate/lock"
)

type Hashed struct {
	candidates map[uint64]string
	hash       func([]byte) [32]byte
	count      uint64
	l          *lock.Lock
}

func New(hash func([]byte) [32]byte) balancer.Balancer {
	return &Hashed{
		candidates: make(map[uint64]string),
		hash:       hash,
		count:      0,
		l:          new(lock.Lock),
	}
}

func (h *Hashed) Add(target string) error {
	h.l.Lock()
	defer h.l.Unlock()

	tombIndex := uint64(0)
	for i := uint64(0); i < h.count; i++ {
		if _, ok := h.candidates[i]; !ok && tombIndex == 0 {
			tombIndex = i
		}
		if h.candidates[i] == target {
			return balancer.ErrorAlreadyExist()
		}
	}

	if tombIndex == 0 {
		h.candidates[h.count] = target
		h.count++
	}

	h.candidates[tombIndex] = target

	return nil
}

func (h *Hashed) Sub(target string) error {
	h.l.Lock()
	defer h.l.Unlock()

	for i := uint64(0); i < h.count; i++ {
		if h.candidates[i] == target {
			delete(h.candidates, i)
			return nil
		}
	}

	return balancer.ErrorValueIsNotExist()
}

func (h *Hashed) Get(id string) (string, error) {
	h.l.Lock()
	defer h.l.Unlock()

	if h.count == 0 {
		return "", balancer.ErrorAnythingNotExist()
	}

	hashed := h.hash([]byte(id))
	i := binary.BigEndian.Uint64(hashed[:8])
	s := i % h.count
	for {
		if _, ok := h.candidates[s]; ok {
			return h.candidates[s], nil
		}
		s++
		if s == h.count {
			s = 0
		}
	}
}

func (h *Hashed) Restore(_ string) error {
	return nil
}
