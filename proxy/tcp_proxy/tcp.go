package tcp_proxy

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/diy-cloud/virtual-gate/lock"
)

type TcpProxy struct {
	upstreamAddress string
	connPool        []*net.TCPConn
	lock            *lock.Lock
}

func NewTcpProxy(upstreamAddress string) *TcpProxy {
	return &TcpProxy{
		upstreamAddress: upstreamAddress,
		connPool:        make([]*net.TCPConn, 0, 10),
		lock:            new(lock.Lock),
	}
}

func (tp *TcpProxy) Connect(client *net.TCPConn) error {
	tp.lock.Lock()
	defer tp.lock.Unlock()
	if len(tp.connPool) == 0 {
		conn, err := net.Dial("tcp", tp.upstreamAddress)
		if err != nil {
			return fmt.Errorf("TcpProxy.Connect: net.Dial: %s", err)
		}
		tcpConn := conn.(*net.TCPConn)
		tcpConn.SetKeepAlive(true)
		tp.connPool = append(tp.connPool, tcpConn)
	}
	conn := tp.connPool[0]
	tp.connPool = tp.connPool[1:]
	go func() {
		upstreamEnd := int64(0)

		buf := [4096]byte{}
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
		tp.connPool = append(tp.connPool, conn)
	}()

	return nil
}

func (tp *TcpProxy) Close() error {
	tp.lock.Lock()
	defer tp.lock.Unlock()
	for i, conn := range tp.connPool {
		if err := conn.Close(); err != nil {
			tp.connPool = tp.connPool[i:]
			return err
		}
	}
	return nil
}

func (tp *TcpProxy) Length() int {
	tp.lock.Lock()
	defer tp.lock.Unlock()
	return len(tp.connPool)
}
