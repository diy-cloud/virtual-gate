package http_proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/diy-cloud/virtual-gate/lock"
	"github.com/diy-cloud/virtual-gate/proxy"
)

type handler struct {
	h func(w http.ResponseWriter, r *http.Request) bool
	l *lock.Lock
}

type proxyMap struct {
	m map[string]*httputil.ReverseProxy
	l *lock.Lock
}

type HttpProxy struct {
	handler  handler
	proxyMap proxyMap
}

func NewHttp() proxy.Proxy {
	hp := new(HttpProxy)

	hp.handler.h = func(w http.ResponseWriter, r *http.Request) bool { return true }
	hp.handler.l = new(lock.Lock)

	hp.proxyMap.m = make(map[string]*httputil.ReverseProxy)
	hp.proxyMap.l = new(lock.Lock)

	return hp
}

func (hp *HttpProxy) SetHandler(h func(w http.ResponseWriter, r *http.Request) bool) {
	hp.handler.l.Lock()
	defer hp.handler.l.Unlock()
	hp.handler.h = h
}

func (hp *HttpProxy) SetUpstreamServer(name string, rawURL string) error {
	url, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("HttpProxy.SetUpstreamServer: url.Parse: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	hp.proxyMap.l.Lock()
	defer hp.proxyMap.l.Unlock()
	if _, ok := hp.proxyMap.m[name]; !ok {
		delete(hp.proxyMap.m, name)
	}
	hp.proxyMap.m[name] = proxy
	return nil
}

func (hp *HttpProxy) ServeHTTP(name string, w http.ResponseWriter, r *http.Request) {
	hp.handler.l.Lock()
	defer hp.handler.l.Unlock()
	if !hp.handler.h(w, r) {
		return
	}

	hp.proxyMap.l.Lock()
	defer hp.proxyMap.l.Unlock()
	if proxy, ok := hp.proxyMap.m[name]; ok {
		proxy.ServeHTTP(w, r)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func (hp *HttpProxy) Serve(address string) error {
	handler := func(w http.ResponseWriter, r *http.Request) {
		hp.ServeHTTP(r.Host, w, r)
	}
	server := http.Server{
		Addr:    address,
		Handler: http.HandlerFunc(handler),
	}
	return server.ListenAndServe()
}
