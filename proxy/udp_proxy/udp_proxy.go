package udp_proxy

import "github.com/diy-cloud/virtual-gate/lock"

type UdpProxy struct {
	lock *lock.Lock
}
