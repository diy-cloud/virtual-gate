package proxy

import (
	"github.com/diy-cloud/virtual-gate/balancer"
	"github.com/diy-cloud/virtual-gate/breaker"
	"github.com/diy-cloud/virtual-gate/limiter"
)

type Proxy interface {
	Serve(address string, limiter limiter.Limiter, acl limiter.Limiter, breaker breaker.CurciutBreaker, balancer balancer.Balancer) error
}
