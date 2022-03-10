package http_proxy

import (
	"fmt"
	"net/http"

	"github.com/diy-cloud/virtual-gate/balancer"
	"github.com/diy-cloud/virtual-gate/breaker"
	"github.com/diy-cloud/virtual-gate/limiter"
	"github.com/diy-cloud/virtual-gate/proxy"
	"golang.org/x/net/http2"
)

type HttpsProxy struct {
	certDir string
	keyDir  string
	proxy   *HttpProxy
}

func NewHttps(cert string, key string, httpProxy *HttpProxy) proxy.Proxy {
	return &HttpsProxy{
		certDir: cert,
		keyDir:  key,
		proxy:   httpProxy,
	}
}

func (hp *HttpsProxy) Serve(address string, limiter limiter.Limiter, acl limiter.Limiter, breaker breaker.Breaker, balancer balancer.Balancer) error {
	handler := func(w http.ResponseWriter, r *http.Request) {
		remote := []byte(r.RemoteAddr)

		for count := 0; count < 10; count++ {
			wr := NewResponse()

			if b, code := limiter.TryTake(remote); !b {
				w.WriteHeader(code)
				return
			}

			if b, code := acl.TryTake(remote); !b {
				w.WriteHeader(code)
				return
			}

			upstreamAddress, err := balancer.Get(r.RemoteAddr)
			if err != nil {
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer balancer.Restore(upstreamAddress)

			if ok := breaker.IsBrokeDown(upstreamAddress); ok {
				continue
			}

			hp.proxy.ServeHTTP(r.Host, wr, r)

			if _, ok := statusCodeSet[wr.StatusCode]; ok {
				breaker.BreakDown(upstreamAddress)
				continue
			}

			breaker.Restore(upstreamAddress)
		}

		w.WriteHeader(http.StatusRequestTimeout)
	}
	server := http.Server{
		Addr:    address,
		Handler: http.HandlerFunc(handler),
	}
	return server.ListenAndServeTLS(hp.certDir, hp.keyDir)
}

type Http2Proxy struct {
	proxy *HttpProxy
}

func NewHttp2(httpProxy *HttpProxy) proxy.Proxy {
	return &Http2Proxy{
		proxy: httpProxy,
	}
}

func (hp *Http2Proxy) Serve(address string, limiter limiter.Limiter, acl limiter.Limiter, breaker breaker.Breaker, balancer balancer.Balancer) error {
	handler := func(w http.ResponseWriter, r *http.Request) {
		remote := []byte(r.RemoteAddr)

		for count := 0; count < 10; count++ {
			wr := NewResponse()

			if b, code := limiter.TryTake(remote); !b {
				w.WriteHeader(code)
				return
			}

			if b, code := acl.TryTake(remote); !b {
				w.WriteHeader(code)
				return
			}

			upstreamAddress, err := balancer.Get(r.RemoteAddr)
			if err != nil {
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer balancer.Restore(upstreamAddress)

			if ok := breaker.IsBrokeDown(upstreamAddress); ok {
				continue
			}

			hp.proxy.ServeHTTP(r.Host, wr, r)

			if _, ok := statusCodeSet[wr.StatusCode]; ok {
				breaker.BreakDown(upstreamAddress)
				continue
			}

			breaker.Restore(upstreamAddress)
		}

		w.WriteHeader(http.StatusRequestTimeout)
	}
	server := http.Server{
		Addr:    address,
		Handler: http.HandlerFunc(handler),
	}
	if err := http2.ConfigureServer(&server, nil); err != nil {
		return fmt.Errorf("Http2Proxy.Serve: http2.ConfigureServer: %w", err)
	}
	return server.ListenAndServeTLS("", "")
}

type Https2Proxy struct {
	certDir string
	keyDir  string
	proxy   *HttpProxy
}

func NewHttp2TLS(cert string, key string, httpProxy *HttpProxy) proxy.Proxy {
	return &Https2Proxy{
		certDir: cert,
		keyDir:  key,
		proxy:   httpProxy,
	}
}

func (hp *Https2Proxy) Serve(address string, limiter limiter.Limiter, acl limiter.Limiter, breaker breaker.Breaker, balancer balancer.Balancer) error {
	handler := func(w http.ResponseWriter, r *http.Request) {
		remote := []byte(r.RemoteAddr)

		for count := 0; count < 10; count++ {
			wr := NewResponse()

			if b, code := limiter.TryTake(remote); !b {
				w.WriteHeader(code)
				return
			}

			if b, code := acl.TryTake(remote); !b {
				w.WriteHeader(code)
				return
			}

			upstreamAddress, err := balancer.Get(r.RemoteAddr)
			if err != nil {
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer balancer.Restore(upstreamAddress)

			if ok := breaker.IsBrokeDown(upstreamAddress); ok {
				continue
			}

			hp.proxy.ServeHTTP(r.Host, wr, r)

			if _, ok := statusCodeSet[wr.StatusCode]; ok {
				breaker.BreakDown(upstreamAddress)
				continue
			}

			breaker.Restore(upstreamAddress)
		}

		w.WriteHeader(http.StatusRequestTimeout)
	}
	server := http.Server{
		Addr:    address,
		Handler: http.HandlerFunc(handler),
	}
	if err := http2.ConfigureServer(&server, nil); err != nil {
		return fmt.Errorf("Https2Proxy.Serve: http2.ConfigureServer: %w", err)
	}
	return server.ListenAndServeTLS(hp.certDir, hp.keyDir)
}
