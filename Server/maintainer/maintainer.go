package maintainer

import (
	"chatserver/reader"
	"chatserver/util"
	"chatserver/writer"
	"fmt"
	"log"
	"time"
)

// ConnMaintainer 用于维护连接, 心跳检测
func ConnMaintainer() {
	ticker := time.NewTicker(util.TimeoutUnit)
	for {
		select {
		case conn := <-util.ConnJoinChan:
			log.Printf("ConnJoinChan...")
			sendMsgChan := make(chan []byte)
			writerStopC, readerStopC, msgSenderStopC, heartSenderStopC := make(chan struct{}, 1),
				make(chan struct{}, 1), make(chan struct{}, 1), make(chan struct{}, 1)
			util.ConnMap[conn] = &util.ConnManager{SendMsgChan: sendMsgChan,
				WriterStopC: writerStopC, ReaderStopC: readerStopC, MsgSenderStopC: msgSenderStopC, HeartSenderStopC: heartSenderStopC}
			go writer.ConnWriter(conn, sendMsgChan, writerStopC)
			go reader.ConnReader(conn, readerStopC)
			go func() {
				welcome := util.Pack{Type: "message", Msg: fmt.Sprintf("欢迎[%v]进入聊天室", conn.RemoteAddr())}
				util.BoardcastMsgChan <- welcome.Marshal()
			}()
		case conn := <-util.UpdateHeartChan:
			log.Printf("UpdateHeartChan...")
			connManager, ok := util.ConnMap[conn]
			if ok {
				connManager.TimeoutCount = 0
			}
		case msg := <-util.BoardcastMsgChan:
			log.Printf("BoardcastMsgChan...")
			for _, v := range util.ConnMap {
				connManager := v
				go func() {
					select {
					case <-connManager.MsgSenderStopC:
					case connManager.SendMsgChan <- msg:
					}
				}()
			}
		case conn := <-util.SendHeartChan:
			log.Printf("SendHeartChan...")
			connManager, ok := util.ConnMap[conn]
			if ok {
				go func() {
					select {
					case <-connManager.HeartSenderStopC:
					case connManager.SendMsgChan <- []byte(`{"type":"heart"}`):
					}
				}()
			}
		case conn := <-util.ConnCloseChan:
			log.Printf("ConnCloseChan...")
			connManager, ok := util.ConnMap[conn]
			if ok {
				remoteAddr := conn.RemoteAddr()
				conn.Close()
				connManager.MsgSenderStopC <- struct{}{}
				connManager.HeartSenderStopC <- struct{}{}
				connManager.WriterStopC <- struct{}{}
				connManager.ReaderStopC <- struct{}{}
				delete(util.ConnMap, conn)
				go func() {
					welcome := util.Pack{Type: "message", Msg: fmt.Sprintf("用户[%v]离开了聊天室", remoteAddr)}
					util.BoardcastMsgChan <- welcome.Marshal()
				}()
			}
		case <-ticker.C:
			log.Printf("Ticker...")
			for conn, connManager := range util.ConnMap {
				connManager.TimeoutCount++
				if connManager.TimeoutCount >= util.TimeoutMax {
					conn.Close()
					connManager.MsgSenderStopC <- struct{}{}
					connManager.HeartSenderStopC <- struct{}{}
					connManager.WriterStopC <- struct{}{}
					connManager.ReaderStopC <- struct{}{}
					delete(util.ConnMap, conn)
				}
			}
		}
	}
}
