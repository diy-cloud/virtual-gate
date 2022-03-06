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
	fmt.Println(len(ups))
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
		errChan := make(chan error)
		clientEnd, upstreamEnd := int64(0), int64(0)

		go func() {
			buf := [4096]byte{}
			err := error(nil)
			for {
				recvN, err := client.Read(buf[:])
				if err != nil {
					err = fmt.Errorf("TcpProxy.Connect: client.Read: %s", err)
					atomic.StoreInt64(&clientEnd, 1)
					break
				}
				sendN, err := conn.Write(buf[:])
				if err != nil {
					err = fmt.Errorf("TcpProxy.Connect: client.Write: %s", err)
					atomic.StoreInt64(&upstreamEnd, 1)
					break
				}
				if recvN != sendN {
					err = fmt.Errorf("TcpProxy.Connect: client.Read != conn.Write: %d != %d", recvN, sendN)
					break
				}
			}
			errChan <- err
		}()

		go func() {
			buf := [4096]byte{}
			err := error(nil)
			for {
				recvN, err := conn.Read(buf[:])
				if err != nil {
					err = fmt.Errorf("TcpProxy.Connect: conn.Read: %s", err)
					atomic.StoreInt64(&upstreamEnd, 1)
					break
				}
				sendN, err := client.Write(buf[:])
				if err != nil {
					err = fmt.Errorf("TcpProxy.Connect: conn.Write: %s", err)
					atomic.StoreInt64(&clientEnd, 1)
					break
				}
				if recvN != sendN {
					err = fmt.Errorf("TcpProxy.Connect: conn.Read != client.Write: %d != %d", recvN, sendN)
					break
				}
			}
			errChan <- err
		}()

		errCount := 0
		for err := range errChan {
			log.Printf("TcpProxy.Connect: %s\n", err)
			errCount++
			if errCount >= 2 {
				close(errChan)
			}
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
