package tcp_proxy

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/diy-cloud/virtual-gate/balancer"
	"github.com/diy-cloud/virtual-gate/breaker"
	"github.com/diy-cloud/virtual-gate/limiter"
	"github.com/diy-cloud/virtual-gate/lock"
	"github.com/diy-cloud/virtual-gate/proxy"
)

type TcpProxy struct {
	connPool map[string][]*net.TCPConn
	lock     *lock.Lock
}

func NewTcpProxy() proxy.Proxy {
	return &TcpProxy{
		connPool: make(map[string][]*net.TCPConn),
		lock:     new(lock.Lock),
	}
}

func (tp *TcpProxy) Connect(client *net.TCPConn, upstreamAddress string) error {
	tp.lock.Lock()
	defer tp.lock.Unlock()
	if _, ok := tp.connPool[upstreamAddress]; !ok {
		tp.connPool[upstreamAddress] = make([]*net.TCPConn, 1)
	}
	if len(tp.connPool[upstreamAddress]) == 0 {
		conn, err := net.Dial("tcp", upstreamAddress)
		if err != nil {
			return fmt.Errorf("TcpProxy.Connect: net.Dial: %s", err)
		}
		tcpConn := conn.(*net.TCPConn)
		tcpConn.SetKeepAlive(true)
		tp.connPool[upstreamAddress] = append(tp.connPool[upstreamAddress], tcpConn)
	}
	conn := tp.connPool[upstreamAddress][0]
	tp.connPool[upstreamAddress] = tp.connPool[upstreamAddress][1:]
	go func() {
		upstreamEnd := int64(0)

		buf := [8192]byte{}
		globalErr := error(nil)
		for {
			recvN, err := client.Read(buf[:])
			if err != nil {
				globalErr = fmt.Errorf("TcpProxy.Connect: client.Read: %s", err)
				break
			}
			sendN, err := conn.Write(buf[:recvN])
			if err != nil {
				globalErr = fmt.Errorf("TcpProxy.Connect: conn.Write: %s", err)
				atomic.StoreInt64(&upstreamEnd, 1)
				break
			}
			if recvN != sendN {
				globalErr = fmt.Errorf("TcpProxy.Connect: client.Read != conn.Write: %d != %d", recvN, sendN)
				break
			}
			recvN, err = conn.Read(buf[:])
			if err != nil {
				globalErr = fmt.Errorf("TcpProxy.Connect: conn.Read: %s", err)
				atomic.StoreInt64(&upstreamEnd, 1)
				break
			}
			sendN, err = client.Write(buf[:recvN])
			if err != nil {
				globalErr = fmt.Errorf("TcpProxy.Connect: client.Write: %s", err)
				break
			}
			if recvN != sendN {
				globalErr = fmt.Errorf("TcpProxy.Connect: conn.Read != client.Write: %d != %d", recvN, sendN)
				break
			}
		}

		if globalErr != nil {
			log.Printf("TcpProxy.Connect: %s\n", globalErr)
		}

		tp.lock.Lock()
		defer tp.lock.Unlock()
		client.Close()
		if atomic.LoadInt64(&upstreamEnd) == 1 {
			conn.Close()
			return
		}
		tp.connPool[upstreamAddress] = append(tp.connPool[upstreamAddress], conn)
	}()

	return nil
}

func (tp *TcpProxy) Close() error {
	tp.lock.Lock()
	defer tp.lock.Unlock()
	for k, pool := range tp.connPool {
		for i, conn := range pool {
			if err := conn.Close(); err != nil {
				tp.connPool[k] = tp.connPool[k][i:]
				return err
			}
		}
	}
	return nil
}

func (tp *TcpProxy) Length() int {
	tp.lock.Lock()
	defer tp.lock.Unlock()
	return len(tp.connPool)
}

func (tp *TcpProxy) Serve(address string, limiter limiter.Limiter, acl limiter.Limiter, breaker breaker.CurciutBreaker, balancer balancer.Balancer) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("TcpProxy.Serve: net.Listen: %s", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("TcpProxy.Serve: listener.Accept: %s", err)
		}
		go func() {
			remote := []byte(conn.RemoteAddr().String())

			for {
				if ok, code := limiter.TryTake(remote); !ok {
					conn.Write([]byte(strconv.Itoa(code)))
					conn.Close()
					return
				}

				if ok, code := acl.TryTake(remote); !ok {
					conn.Write([]byte(strconv.Itoa(code)))
					conn.Close()
					return
				}

				upstreamAddress, err := balancer.Get(conn.RemoteAddr().String())
				if err != nil {
					conn.Write([]byte(err.Error()))
					conn.Close()
					return
				}
				defer balancer.Restore(upstreamAddress)

				if ok := breaker.IsBrokeDown(upstreamAddress); !ok {
					continue
				}

				if err := tp.Connect(conn.(*net.TCPConn), upstreamAddress); err != nil {
					if err != io.EOF {
						breaker.BreakDown(upstreamAddress)
						conn.Write([]byte(err.Error()))
						conn.Close()
						return
					}
					conn.Close()
				}

				breaker.Restore(upstreamAddress)
			}
		}()
	}
}
