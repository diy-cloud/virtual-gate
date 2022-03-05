package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/diy-cloud/virtual-gate/limiter/slide_count"
)

var maxConn = int64(30000)
var maxReq = 100000

func main() {
	rs := time.Microsecond * 10
	for i := 0; i < 100; i++ {
		r := testA()
		rs = (rs + r) / 2
	}
	fmt.Println(rs)
}

func testA() time.Duration {
	limiter := slide_count.New(float64(maxConn), time.Microsecond)
	key := []byte("localhost")
	c := int64(0)
	wg := new(sync.WaitGroup)
	m := new(sync.Mutex)
	l := []time.Duration{}
	for i := 0; i < maxReq; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s := time.Now()
			for {
				b, _ := limiter.TryTake(key)
				if b {
					atomic.AddInt64(&c, 1)
					break
				}
			}
			e := time.Now()
			m.Lock()
			l = append(l, e.Sub(s))
			m.Unlock()
		}()
	}
	wg.Wait()
	_ = c
	rs := l[0]
	for _, v := range l[1:] {
		rs = (rs + v) / 2
	}
	return rs
}
