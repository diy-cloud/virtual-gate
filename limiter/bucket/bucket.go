package bucket

import (
	"net/http"
	"time"

	"github.com/diy-cloud/virtual-gate/limiter"
	"github.com/diy-cloud/virtual-gate/lock"
)

type Bucket struct {
	lock      *lock.Lock
	maxTokens int64
	tokens    int64
	regenTime time.Duration
	ch        chan struct{}
}

func New(maxTokens, regenPerSecond int) limiter.Limiter {
	ch := make(chan struct{}, maxTokens)
	bucket := &Bucket{
		lock:      new(lock.Lock),
		maxTokens: int64(maxTokens),
		tokens:    int64(maxTokens),
		regenTime: time.Second / time.Duration(regenPerSecond),
		ch:        ch,
	}
	go func() {
		for range ch {
			time.Sleep(bucket.regenTime)
			bucket.lock.Lock()
			bucket.tokens++
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
	b.ch <- struct{}{}
	return true, http.StatusOK
}

func (b *Bucket) TryTake(_ []byte) (bool, int) {
	if passed, code := b.decreaseToken(); !passed {
		return passed, code
	}
	return true, http.StatusOK
}
