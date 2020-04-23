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

// ConnReader 用于从conn读消息
func ConnReader(conn *net.TCPConn, stopC <-chan struct{}) {
	log.Printf("ConnReader is running...")
	headBuf := make([]byte, 4)
	for {
		// 先读包头
		conn.SetReadDeadline(time.Now().Add(time.Duration(util.TimeoutMax) * util.TimeoutUnit))
		_, err := io.ReadFull(conn, headBuf)
		if err != nil {
			// TODO ConnCloseChan阻塞效率不高
			util.ConnCloseChan <- conn
			return
		}
		// TODO UpdateHeartChan阻塞效率不高
		select {
		case <-stopC:
			return
		case util.UpdateHeartChan <- conn:
		}

		// 解析包头
		packLen := binary.LittleEndian.Uint32(headBuf)
		if packLen == 0 {
			// TODO ConnCloseChan阻塞效率不高
			util.ConnCloseChan <- conn
			return
		}

		// 读取包本体
		bufPack := make([]byte, packLen)
		conn.SetReadDeadline(time.Now().Add(time.Duration(util.TimeoutMax) * util.TimeoutUnit))
		_, err = io.ReadFull(conn, bufPack)
		if err != nil {
			// TODO ConnCloseChan塞效率不高
			util.ConnCloseChan <- conn
			return
		}
		// TODO UpdateHeartChan阻塞效率不高
		select {
		case <-stopC:
			return
		case util.UpdateHeartChan <- conn:
		}

		// 解析包
		pack := &util.Pack{}
		err = pack.Unmarshal(bufPack)
		if err != nil {
			// TODO ConnCloseChan阻塞效率不高
			util.ConnCloseChan <- conn
			return
		}

		// 处理包
		switch {
		case pack.Type == "heart":
			// TODO SendHeartChan阻塞效率不高
			util.SendHeartChan <- conn
			continue
		case pack.Type == "message":
			if pack.Msg == "" {
				// TODO ConnCloseChan阻塞效率不高
				util.ConnCloseChan <- conn
				return
			}
			sendPack := &util.Pack{Type: "message", Msg: fmt.Sprintf("[%v] %v", conn.RemoteAddr().String(), pack.Msg)}
			util.BoardcastMsgChan <- sendPack.Marshal()
		default:
			// TODO ConnCloseChan阻塞效率不高
			util.ConnCloseChan <- conn
			return
		}
	}
}
