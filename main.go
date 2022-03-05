package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/diy-cloud/virtual-gate/limiter/slide"
)

func main() {
	limiter := slide.New(30000)
	key := []byte("localhost")
	s := time.Now()
	c := int64(0)
	wg := new(sync.WaitGroup)
	for i := 0; i < 1000000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				b, _ := limiter.TryTake(key)
				if b {
					atomic.AddInt64(&c, 1)
					break
				}
			}
		}()
	}
	wg.Wait()
	e := time.Now()
	fmt.Println(e.Sub(s))
	fmt.Println(c)
}
