package slide

import (
	"sync"
	"time"

	"github.com/diy-cloud/virtual-gate/limiter"
	"github.com/diy-cloud/virtual-gate/lock"
)

var timeNodePool = sync.Pool{
	New: func() interface{} {
		return new(timeNode)
	},
}

type timeNode struct {
	value int64
	next  *timeNode
}

type Slide struct {
	recentlyTakensSet map[string]int64
	lock              *lock.Lock
	maxConnPerSecond  float64
	timeNodeHead      *timeNode
	timeNodeTail      *timeNode
	timeNodeCount     int64
}

func New(maxConnPerSecond float64) limiter.Limiter {
	return &Slide{
		recentlyTakensSet: make(map[string]int64),
		lock:              new(lock.Lock),
		maxConnPerSecond:  maxConnPerSecond,
		timeNodeHead:      nil,
		timeNodeTail:      nil,
		timeNodeCount:     0,
	}
}

func (s *Slide) TryTake(key []byte) (bool, int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	now := time.Now().UnixMicro()
	past := now - 1000000
	for cur := s.timeNodeHead; cur != nil; cur = cur.next {
		if cur.value < past {
			if cur.next == nil {
				s.timeNodeHead = nil
				s.timeNodeTail = nil
			}
			cur.next = nil
			timeNodePool.Put(cur)
			s.timeNodeCount--
		} else {
			break
		}
	}
	if s.timeNodeCount >= int64(s.maxConnPerSecond) {
		return false, 0
	}

	newTimeNode := timeNodePool.Get().(*timeNode)
	newTimeNode.value = now
	newTimeNode.next = nil
	if s.timeNodeHead == nil {
		s.timeNodeHead = newTimeNode
		s.timeNodeTail = newTimeNode
	} else {
		s.timeNodeTail.next = newTimeNode
		s.timeNodeTail = newTimeNode
	}
	s.timeNodeCount++
	return true, 0
}
