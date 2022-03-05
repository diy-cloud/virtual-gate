package bucket

import (
	"encoding/hex"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/diy-cloud/virtual-gate/limiter"
)

var nodePool = sync.Pool{
	New: func() interface{} {
		return new(node)
	},
}

type node struct {
	value string
	next  *node
}

type lock struct {
	i int64
}

func (l *lock) Lock() {
	for !atomic.CompareAndSwapInt64(&l.i, 0, 1) {
		runtime.Gosched()
	}
}
func (l *lock) Unlock() {
	atomic.StoreInt64(&l.i, 0)
}

type Bucket struct {
	recentlyTakensHead    *node
	recentlyTakensTail    *node
	recentlyTakensSet     map[string]int64
	lock                  *lock
	maxTokens             int64
	tokens                int64
	regenTime             time.Duration
	sameRemoteIPLimitRate float64
	goPool                chan struct{}
}

func New(maxTokens, regenPerSecond int, sameRemoteIPLimitRate float64) limiter.Limiter {
	ch := make(chan struct{}, maxTokens)
	bucket := &Bucket{
		recentlyTakensHead:    nil,
		recentlyTakensTail:    nil,
		recentlyTakensSet:     make(map[string]int64),
		lock:                  new(lock),
		maxTokens:             int64(maxTokens),
		tokens:                int64(maxTokens),
		regenTime:             time.Second / time.Duration(regenPerSecond),
		sameRemoteIPLimitRate: sameRemoteIPLimitRate,
		goPool:                ch,
	}
	go func() {
		for range ch {
			time.Sleep(bucket.regenTime)
			bucket.lock.Lock()
			bucket.tokens++
			if bucket.recentlyTakensHead == nil {
				bucket.lock.Unlock()
				continue
			}
			node := bucket.recentlyTakensHead
			bucket.recentlyTakensHead = node.next
			if bucket.recentlyTakensHead == nil {
				bucket.recentlyTakensTail = nil
			}
			bucket.recentlyTakensSet[node.value]--
			nodePool.Put(node)
			bucket.lock.Unlock()
		}
	}()
	return bucket
}

func (b *Bucket) decreaseToken() (bool, int) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.tokens <= 0 {
		return false, http.StatusNotAcceptable
	}
	b.tokens--
	b.goPool <- struct{}{}
	return true, http.StatusOK
}

func (b *Bucket) checkTaken(key string) (bool, int) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.recentlyTakensSet[key] >= int64(float64(b.maxTokens)*b.sameRemoteIPLimitRate) {
		return false, http.StatusTooManyRequests
	}
	return true, http.StatusOK
}

func (b *Bucket) appendTaken(key string) (bool, int) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.recentlyTakensSet[key]++
	newNode := nodePool.Get().(*node)
	newNode.value = key
	newNode.next = nil
	if b.recentlyTakensHead == nil {
		b.recentlyTakensHead = newNode
		b.recentlyTakensTail = newNode
		return true, http.StatusOK
	}
	b.recentlyTakensTail.next = newNode
	b.recentlyTakensTail = newNode
	return true, http.StatusOK
}

func (b *Bucket) TryTake(key []byte) (bool, int) {
	hexKey := hex.EncodeToString(key)
	if passed, code := b.checkTaken(hexKey); !passed {
		return passed, code
	}
	if passed, code := b.decreaseToken(); !passed {
		return passed, code
	}
	if passed, code := b.appendTaken(hexKey); !passed {
		return passed, code
	}
	return true, http.StatusOK
}
