package main

import (
	"fmt"
	"hash/fnv"
	"log"
	"sync"
	"time"

	"github.com/diy-cloud/virtual-gate/balancer/hashed"
	"github.com/diy-cloud/virtual-gate/lock"
)

func main() {
	balancer := hashed.New(fnv.New64a())

	ids := []string{
		"localhost:8080",
		"localhost:8081",
		"localhost:8082",
		"localhost:8083",
		"localhost:8084",
		"localhost:8085",
		"localhost:8086",
		"localhost:8087",
		"localhost:8088",
		"localhost:8089",
		"localhost:8090",
		"localhost:8091",
		"localhost:8092",
	}

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
