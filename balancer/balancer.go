package balancer

type Balancer interface {
	Get() (string, error)
}
