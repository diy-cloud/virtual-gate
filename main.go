package main

import (
	"fmt"
	"math/rand"

	"github.com/diy-cloud/virtual-gate/breaker/count_breaker"
)

func main() {
	l := make([]int, 1000)
	for k := 0; k < 1000; k++ {
		breaker := count_breaker.New(100, 20)
		brokenCount := 0
		for i := 0; i < 1000; i++ {
			if breaker.IsBrokeDown("test") {
				brokenCount++
				i--
				continue
			}
			if rand.Int63n(100) < 80 {
				breaker.BreakDown("test")
				continue
			}
			breaker.Restore("test")
		}
		l[k] = brokenCount
	}
	avg := 0
	for _, v := range l {
		avg += v
	}
	avg = avg / len(l)
	fmt.Println(avg)
}
