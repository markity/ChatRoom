package main

import (
	"chatserver/maintainer"
	"chatserver/util"
	"log"
	"net"
)

var serverAddr = "127.0.0.1:8000"

func main() {
	addr, err := net.ResolveTCPAddr("tcp4", serverAddr)
	if err != nil {
		log.Fatalf("failed to ResolveTCPAddr: %v", err)
	}

	listener, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		log.Fatalf("failed to ListenTCP: %v", err)
	}

	log.Printf("server started at %v...", serverAddr)

	go maintainer.ConnMaintainer()
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("failed to AcceptTCP: %v", err)
			continue
		}

		// TODO 阻塞效率不高
		util.ConnJoinChan <- conn
	}
}
