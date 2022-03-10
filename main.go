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
	breaker := count_breaker.New(8, 10)
	balancer := least.New()
	proxy := http_proxy.NewHttp()

	if err := proxy.Serve("0.0.0.0:80", limiter, acl, breaker, balancer); err != nil {
		panic(err)
	}
}
