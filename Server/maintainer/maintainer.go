package maintainer

import (
	"chatserver/reader"
	"chatserver/util"
	"chatserver/writer"
	"fmt"
	"log"
	"net"
	"time"
)

type connManager struct {
	// 用于向connWriter写入消息
	SendMsgChan chan []byte
	// 用于控制线程停止
	WriterStopC      chan struct{}
	ReaderStopC      chan struct{}
	MsgSenderStopC   chan struct{}
	HeartSenderStopC chan struct{}
	// 记录超时次数
	TimeoutCount int
}

// ConnMaintainer 用于维护连接, 心跳检测
func ConnMaintainer() {
	connMap := make(map[*net.TCPConn]*connManager)
	ticker := time.NewTicker(util.TimeoutUnit)
	for {
		select {
		case conn := <-util.ConnJoinChan:
			log.Printf("ConnJoinChan...")
			sendMsgChan := make(chan []byte)
			writerStopC, readerStopC, msgSenderStopC, heartSenderStopC := make(chan struct{}, 1),
				make(chan struct{}, 1), make(chan struct{}, 1), make(chan struct{}, 1)
			connMap[conn] = &connManager{SendMsgChan: sendMsgChan,
				WriterStopC: writerStopC, ReaderStopC: readerStopC, MsgSenderStopC: msgSenderStopC, HeartSenderStopC: heartSenderStopC}
			go writer.ConnWriter(conn, sendMsgChan, writerStopC)
			go reader.ConnReader(conn, readerStopC)
			go func() {
				welcome := util.Pack{Type: "message", Msg: fmt.Sprintf("欢迎[%v]进入聊天室", conn.RemoteAddr())}
				util.BoardcastMsgChan <- welcome.Marshal()
			}()
		case conn := <-util.UpdateHeartChan:
			log.Printf("UpdateHeartChan...")
			cm, ok := connMap[conn]
			if ok {
				cm.TimeoutCount = 0
			}
		case msg := <-util.BoardcastMsgChan:
			log.Printf("BoardcastMsgChan...")
			for _, v := range connMap {
				cm := v
				go func() {
					select {
					case <-cm.MsgSenderStopC:
					case cm.SendMsgChan <- msg:
					}
				}()
			}
		case conn := <-util.SendHeartChan:
			log.Printf("SendHeartChan...")
			cm, ok := connMap[conn]
			if ok {
				go func() {
					select {
					case <-cm.HeartSenderStopC:
					case cm.SendMsgChan <- []byte(`{"type":"heart"}`):
					}
				}()
			}
		case conn := <-util.ConnCloseChan:
			log.Printf("ConnCloseChan...")
			cm, ok := connMap[conn]
			if ok {
				remoteAddr := conn.RemoteAddr()
				conn.Close()
				cm.MsgSenderStopC <- struct{}{}
				cm.HeartSenderStopC <- struct{}{}
				cm.WriterStopC <- struct{}{}
				cm.ReaderStopC <- struct{}{}
				delete(connMap, conn)
				go func() {
					welcome := util.Pack{Type: "message", Msg: fmt.Sprintf("用户[%v]离开了聊天室", remoteAddr)}
					util.BoardcastMsgChan <- welcome.Marshal()
				}()
			}
		case <-ticker.C:
			log.Printf("Ticker...")
			for conn, cm := range connMap {
				cm.TimeoutCount++
				if cm.TimeoutCount >= util.TimeoutMax {
					conn.Close()
					cm.MsgSenderStopC <- struct{}{}
					cm.HeartSenderStopC <- struct{}{}
					cm.WriterStopC <- struct{}{}
					cm.ReaderStopC <- struct{}{}
					delete(connMap, conn)
				}
			}
		}
	}
}
