package tcp_proxy

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/diy-cloud/virtual-gate/lock"
)

type TcpProxy struct {
	connPool map[string][]*net.TCPConn
	lock     *lock.Lock
}

func NewTcpProxy() *TcpProxy {
	return &TcpProxy{
		connPool: make(map[string][]*net.TCPConn),
		lock:     new(lock.Lock),
	}
}

func (tp *TcpProxy) Connect(upstreamAddress string, client *net.TCPConn) error {
	tp.lock.Lock()
	defer tp.lock.Unlock()
	ups, ok := tp.connPool[upstreamAddress]
	if !ok {
		ups = make([]*net.TCPConn, 0)
	}
	if len(ups) == 0 {
		conn, err := net.Dial("tcp", upstreamAddress)
		if err != nil {
			return fmt.Errorf("TcpProxy.Connect: net.Dial: %s", err)
		}
		tcpConn := conn.(*net.TCPConn)
		tcpConn.SetKeepAlive(true)
		ups = append(ups, tcpConn)
	}
	conn := ups[0]
	ups = ups[1:]
	tp.connPool[upstreamAddress] = ups
	go func() {
		clientEnd, upstreamEnd := int64(0), int64(0)

		buf := [4096]byte{}
		globalErr := error(nil)
		for {
			recvN, err := client.Read(buf[:])
			if err != nil {
				globalErr = fmt.Errorf("TcpProxy.Connect: client.Read: %s", err)
				atomic.StoreInt64(&clientEnd, 1)
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
				atomic.StoreInt64(&clientEnd, 1)
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
		if atomic.LoadInt64(&clientEnd) == 1 {
			client.Close()
		}
		if atomic.LoadInt64(&upstreamEnd) == 1 {
			conn.Close()
			return
		}
		ups, ok := tp.connPool[upstreamAddress]
		if !ok {
			ups = make([]*net.TCPConn, 0, 1)
		}
		ups = append(ups, conn)
		tp.connPool[upstreamAddress] = ups
	}()

	return nil
}

func (tp *TcpProxy) Close() error {
	tp.lock.Lock()
	defer tp.lock.Unlock()
	for name, ups := range tp.connPool {
		for j, conn := range ups {
			if err := conn.Close(); err != nil {
				tp.connPool[name] = ups[j:]
				return err
			}
		}
	}
	return nil
}
