package main

import (
	"fmt"

	"github.com/diy-cloud/virtual-gate/breaker/count_breaker"
)

func main() {
	breaker := count_breaker.New(10)
	brokenCount := 0
	for i := 0; i < 100; i++ {
		if breaker.IsBrokeDown("test") {
			brokenCount++
			i--
		}
		if i%3 == 0 {
			breaker.BreakDown("test")
			continue
		}
		breaker.Restore("test")
	}
	fmt.Println(brokenCount)
}
