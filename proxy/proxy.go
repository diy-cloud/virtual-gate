package proxy

type Proxy interface {
	Serve(address string) error
}
