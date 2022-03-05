package main

import (
	"log"
	"sync"
	"time"

	"github.com/diy-cloud/virtual-gate/limiter/slide_count"
)

func main() {
	limiter := slide_count.New(5, time.Second)
	wg := new(sync.WaitGroup)
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Millisecond * 500)
			for {
				b, _ := limiter.TryTake(nil)
				if b {
					log.Println("take")
					break
				}
			}
		}()
	}
	wg.Wait()
}
