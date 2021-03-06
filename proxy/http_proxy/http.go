package http_proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/diy-cloud/virtual-gate/balancer"
	"github.com/diy-cloud/virtual-gate/breaker"
	"github.com/diy-cloud/virtual-gate/limiter"
	"github.com/diy-cloud/virtual-gate/lock"
)

var statusCodeSet = map[int]struct{}{
	http.StatusNotFound:              {},
	http.StatusRequestTimeout:        {},
	http.StatusFailedDependency:      {},
	http.StatusInternalServerError:   {},
	http.StatusBadGateway:            {},
	http.StatusServiceUnavailable:    {},
	http.StatusGatewayTimeout:        {},
	http.StatusVariantAlsoNegotiates: {},
	http.StatusInsufficientStorage:   {},
	http.StatusLoopDetected:          {},
	http.StatusNotExtended:           {},
}

type HttpProxy struct {
	proxyCache map[string]*httputil.ReverseProxy
	l          *lock.Lock
}

func NewHttp() *HttpProxy {
	return &HttpProxy{
		proxyCache: make(map[string]*httputil.ReverseProxy),
		l:          new(lock.Lock),
	}
}

func (hp *HttpProxy) ServeHTTP(name string, w http.ResponseWriter, r *http.Request) {
	var upstreamServer *httputil.ReverseProxy
	hp.l.Lock()
	if l, ok := hp.proxyCache[name]; ok {
		hp.proxyCache[name] = l
	}
	if upstreamServer == nil {
		url, err := url.Parse(name)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		upstreamServer = httputil.NewSingleHostReverseProxy(url)
	}
	hp.l.Unlock()

	upstreamServer.ServeHTTP(w, r)

	hp.l.Lock()
	if _, ok := hp.proxyCache[name]; !ok {
		hp.proxyCache[name] = upstreamServer
	}
	hp.l.Unlock()
}

func (hp *HttpProxy) Serve(address string, limiter limiter.Limiter, acc limiter.Limiter, breaker breaker.Breaker, balancer balancer.Balancer) error {
	handler := func(w http.ResponseWriter, r *http.Request) {
		remote := []byte(r.RemoteAddr)

		wr := NewResponse()

		if b, code := limiter.TryTake(remote); !b {
			log.Println("HttpProxy.Serve: limiter.TryTake: false from", r.RemoteAddr)
			w.WriteHeader(code)
			return
		}

		if b, code := acc.TryTake(remote); !b {
			log.Println("HttpProxy.Serve: acl.TryTake: false from", r.RemoteAddr)
			w.WriteHeader(code)
			return
		}

		for count := 0; count < 10; count++ {
			upstreamAddress, err := balancer.Get(r.RemoteAddr)
			if err != nil {
				log.Println("HttpProxy.Serve: balancer.Get:", err, "from", r.RemoteAddr)
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer balancer.Restore(upstreamAddress)

			if ok := breaker.IsBrokeDown(upstreamAddress); ok {
				log.Println("HttpProxy.Serve: breaker.IsBrokeDown: true from", r.RemoteAddr, "to", upstreamAddress)
				continue
			}

			hp.ServeHTTP(upstreamAddress, wr, r)

			if _, ok := statusCodeSet[wr.StatusCode]; ok {
				log.Println("HttpProxy.Serve: breakDown: true from", r.RemoteAddr, "to", upstreamAddress)
				breaker.BreakDown(upstreamAddress)
				continue
			}

			breaker.Restore(upstreamAddress)

			if _, ok := statusCodeSet[wr.StatusCode]; !ok {
				w.Write(wr.Body)
				w.WriteHeader(wr.StatusCode)
				break
			}
		}
	}
	server := http.Server{
		Addr:    address,
		Handler: http.HandlerFunc(handler),
	}
	return server.ListenAndServe()
}
