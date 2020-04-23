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

	go maintainer.ConnMaintainer()
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("failed to AcceptTCP: %v", err)
			continue
		}
		util.ConnJoinChan <- conn
	}
}
