package main

import (
	"fmt"
	"log"
	"net"

	"github.com/diy-cloud/virtual-gate/proxy/tcp_proxy"
)

func main() {
	tcpProxy := tcp_proxy.NewTcpProxy()

	lis, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 8999,
	})
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := lis.AcceptTCP()
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			if err := tcpProxy.Connect("localhost:8080", conn); err != nil {
				log.Println(err)
			}
			fmt.Println(tcpProxy.Length("localhost:8080"))
		}()
	}
}
