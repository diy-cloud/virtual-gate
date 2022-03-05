package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/diy-cloud/virtual-gate/limiter/slide_count"
	"github.com/diy-cloud/virtual-gate/limiter/slide_count/acl"
)

func main() {
	limiter := slide_count.New(20000, time.Microsecond)
	acl := acl.New(2000, time.Microsecond)
	wg := new(sync.WaitGroup)
	s := time.Now()
	for i := 0; i < 10000000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				b, _ := limiter.TryTake(nil)
				a, _ := acl.TryTake([]byte("test"))
				if b && a {
					break
				}
			}
		}()
	}
	wg.Wait()
	e := time.Now()
	fmt.Println(e.Sub(s))
}
