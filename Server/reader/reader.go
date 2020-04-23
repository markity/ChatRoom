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
		// 获取packLen
		conn.SetReadDeadline(time.Now().Add(time.Duration(util.TimeoutMax) * util.TimeoutUnit))
		_, err := io.ReadFull(conn, headBuf)
		if err != nil {
			util.ConnCloseChan <- conn
			return
		}
		select {
		case <-stopC:
			return
		case util.UpdateHeartChan <- conn:
		}
		packLen := binary.LittleEndian.Uint32(headBuf)
		if packLen == 0 {
			util.ConnCloseChan <- conn
			return
		}

		// 读取包本体
		bufPack := make([]byte, packLen)
		conn.SetReadDeadline(time.Now().Add(time.Duration(util.TimeoutMax) * util.TimeoutUnit))
		_, err = io.ReadFull(conn, bufPack)
		if err != nil {
			util.ConnCloseChan <- conn
			return
		}
		select {
		case <-stopC:
			return
		case util.UpdateHeartChan <- conn:
		}

		// 解析包
		pack := &util.Pack{}
		err = pack.Unmarshal(bufPack)
		if err != nil {
			util.ConnCloseChan <- conn
			return
		}

		// 处理包
		switch {
		case pack.Type == "heart":
			util.SendHeartChan <- conn
		case pack.Type == "message":
			if pack.Msg == "" {
				util.ConnCloseChan <- conn
				return
			}
			sendPack := &util.Pack{Type: "message", Msg: fmt.Sprintf("[%v] [%v] %v",
				conn.RemoteAddr().String(), time.Now().Format("2006-01-02 15:04"), pack.Msg)}
			util.BoardcastMsgChan <- sendPack.Marshal()
		default:
			util.ConnCloseChan <- conn
			return
		}
	}
}
