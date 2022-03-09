package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/diy-cloud/virtual-gate/balancer/round"
	"github.com/diy-cloud/virtual-gate/lock"
)

func main() {
	balancer := round.New()

	ids := []string{"a", "b", "c", "d", "e", "d", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

	balancer.Add("a")
	balancer.Add("b")
	balancer.Add("c")
	balancer.Add("d")
	balancer.Add("e")

	used := map[string]int{}
	lock := new(lock.Lock)

	wg := new(sync.WaitGroup)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			target, err := balancer.Get(id)
			if err != nil {
				log.Println(err)
				return
			}
			lock.Lock()
			used[target]++
			lock.Unlock()
			time.Sleep(time.Millisecond * 100)
			balancer.Restore(target)
		}(ids[i%len(ids)])
	}
	wg.Wait()

	fmt.Println(used)
}
