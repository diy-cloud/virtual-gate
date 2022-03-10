package main

import (
	"time"

	"github.com/diy-cloud/virtual-gate/balancer/least"
	"github.com/diy-cloud/virtual-gate/breaker/count_breaker"
	"github.com/diy-cloud/virtual-gate/limiter/slide_count"
	"github.com/diy-cloud/virtual-gate/limiter/slide_count/acc"
	"github.com/diy-cloud/virtual-gate/proxy/http_proxy"
)

func main() {
	limiter := slide_count.New(30000, time.Microsecond)
	acc := acc.New(60, time.Microsecond)
	breaker := count_breaker.New(200, 10)
	balancer := least.New()
	balancer.Add("http://127.0.0.1:8080")
	balancer.Add("http://127.0.0.1:8081")
	balancer.Add("http://127.0.0.1:8082")
	balancer.Add("http://127.0.0.1:8083")
	balancer.Add("http://127.0.0.1:8084")
	balancer.Add("http://127.0.0.1:8085")
	balancer.Add("http://127.0.0.1:8086")
	balancer.Add("http://127.0.0.1:8087")
	balancer.Add("http://127.0.0.1:8088")
	balancer.Add("http://127.0.0.1:8089")
	proxy := http_proxy.NewHttp()

	if err := proxy.Serve("0.0.0.0:9999", limiter, acc, breaker, balancer); err != nil {
		panic(err)
	}
}
