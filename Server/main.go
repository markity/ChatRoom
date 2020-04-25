package main

import (
	"chatserver/maintainer"
	"chatserver/util"
	"log"
	"net"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp4", util.ServerAddr)
	if err != nil {
		log.Fatalf("failed to ResolveTCPAddr: %v", err)
	}
	listener, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		log.Fatalf("failed to ListenTCP: %v", err)
	}
	log.Printf("server started at %v...", util.ServerAddr)

	connJoinC := make(chan *net.TCPConn, util.ChanCap)
	// 开启线程维护器
	maintainer.NewMaintainer(connJoinC).Start()

	// 阻塞接收新的连接
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("failed to AcceptTCP: %v", err)
			continue
		}
		connJoinC <- conn
	}
}
