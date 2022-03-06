package slide_log

import (
	"sync"
	"time"

	"github.com/diy-cloud/virtual-gate/limiter"
	"github.com/diy-cloud/virtual-gate/lock"
)

var timeNodePool = sync.Pool{
	New: func() any {
		return new(timeNode)
	},
}

type timeNode struct {
	value int64
	next  *timeNode
}

type SlideLog struct {
	recentlyTakensSet map[string]int64
	lock              *lock.Lock
	maxConnPerSecond  int64
	timeNodeHead      *timeNode
	timeNodeTail      *timeNode
	timeNodeCount     int64
}

func New(maxConnPerMicrosecond int64) limiter.Limiter {
	return &SlideLog{
		recentlyTakensSet: make(map[string]int64),
		lock:              new(lock.Lock),
		maxConnPerSecond:  maxConnPerMicrosecond,
		timeNodeHead:      nil,
		timeNodeTail:      nil,
		timeNodeCount:     0,
	}
}

func (s *SlideLog) TryTake(_ []byte) (bool, int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	now := time.Now().UnixNano()
	past := now - 1000
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
	if s.timeNodeCount >= s.maxConnPerSecond {
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
