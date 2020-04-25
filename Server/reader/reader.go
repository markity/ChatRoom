package reader

import (
	"chatserver/util"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

func NewReader(conn *net.TCPConn, broadcastC chan []byte, updateHeartC chan *net.TCPConn,
	sendHeartC chan *net.TCPConn, connCloseC chan *net.TCPConn, stopC chan struct{}) *Reader {
	return &Reader{Conn: conn, BroadcastC: broadcastC, UpdateHeartC: updateHeartC,
		SendHeartC: sendHeartC, ConnCloseC: connCloseC, StopC: stopC}
}

type Reader struct {
	Conn *net.TCPConn

	BroadcastC   chan []byte
	UpdateHeartC chan *net.TCPConn
	SendHeartC   chan *net.TCPConn
	ConnCloseC   chan *net.TCPConn
	StopC        chan struct{}
}

func (r *Reader) Start() {
	go func() {
		log.Printf("Reader is running...")
		headBuf := make([]byte, 4)
		for {
			// 获取packLen
			err := r.Conn.SetReadDeadline(time.Now().Add(time.Duration(util.TimeoutMax) * util.TimeoutUnit))
			if err != nil {
				r.ConnCloseC <- r.Conn
				return
			}
			_, err = io.ReadFull(r.Conn, headBuf)
			if err != nil {
				r.ConnCloseC <- r.Conn
				return
			}
			select {
			case <-r.StopC:
				return
			case r.UpdateHeartC <- r.Conn:
			}
			packLen := binary.LittleEndian.Uint32(headBuf)
			if packLen == 0 {
				r.ConnCloseC <- r.Conn
				return
			}

			// 读取包本体
			bufPack := make([]byte, packLen)
			err = r.Conn.SetReadDeadline(time.Now().Add(time.Duration(util.TimeoutMax) * util.TimeoutUnit))
			if err != nil {
				r.ConnCloseC <- r.Conn
				return
			}
			_, err = io.ReadFull(r.Conn, bufPack)
			if err != nil {
				r.ConnCloseC <- r.Conn
				return
			}
			select {
			case <-r.StopC:
				return
			case r.UpdateHeartC <- r.Conn:
			}

			// 解析包
			pack := &util.Pack{}
			err = pack.Unmarshal(bufPack)
			if err != nil {
				r.ConnCloseC <- r.Conn
				return
			}

			// 处理包
			switch {
			case pack.Type == "heart":
				r.SendHeartC <- r.Conn
			case pack.Type == "message":
				if pack.Msg == "" {
					r.ConnCloseC <- r.Conn
					return
				}
				sendPack := &util.Pack{Type: "message", Msg: fmt.Sprintf("[%v] [%v] %v",
					r.Conn.RemoteAddr().String(), time.Now().Format("2006-01-02 15:04"), pack.Msg)}
				r.BroadcastC <- sendPack.Marshal()
			default:
				r.ConnCloseC <- r.Conn
				return
			}
		}
	}()
}
