package main

import (
	"time"

	"github.com/diy-cloud/virtual-gate/balancer/least"
	"github.com/diy-cloud/virtual-gate/breaker/count_breaker"
	"github.com/diy-cloud/virtual-gate/limiter/slide_count"
	"github.com/diy-cloud/virtual-gate/limiter/slide_count/acl"
	"github.com/diy-cloud/virtual-gate/proxy/http_proxy"
)

func main() {
	limiter := slide_count.New(30000, time.Microsecond)
	acl := acl.New(60, time.Microsecond)
	breaker := count_breaker.New(200, 10)
	balancer := least.New()
	balancer.Add("localhost:8080")
	balancer.Add("localhost:8081")
	balancer.Add("localhost:8082")
	balancer.Add("localhost:8083")
	balancer.Add("localhost:8084")
	balancer.Add("localhost:8085")
	balancer.Add("localhost:8086")
	balancer.Add("localhost:8087")
	balancer.Add("localhost:8088")
	balancer.Add("localhost:8089")
	proxy := http_proxy.NewHttp()

	if err := proxy.Serve("0.0.0.0:9999", limiter, acl, breaker, balancer); err != nil {
		panic(err)
	}
}
