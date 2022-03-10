package http_proxy

import (
	"fmt"
	"net/http"

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

func (hp *HttpsProxy) Serve(address string) error {
	handler := func(w http.ResponseWriter, r *http.Request) {
		hp.proxy.ServeHTTP(r.Host, w, r)
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

func (h *Http2Proxy) Serve(address string) error {
	handler := func(w http.ResponseWriter, r *http.Request) {
		h.proxy.ServeHTTP(r.Host, w, r)
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

func (hp *Https2Proxy) Serve(address string) error {
	handler := func(w http.ResponseWriter, r *http.Request) {
		hp.proxy.ServeHTTP(r.Host, w, r)
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
