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

func NewMaintainer(connJoinC chan *net.TCPConn) *Maintainer {
	return &Maintainer{ConnJoinC: connJoinC}
}

type Maintainer struct {
	ConnJoinC chan *net.TCPConn
}

func (m *Maintainer) Start() {
	go func() {
		log.Printf("Maintainer is running...")

		connMap := make(map[*net.TCPConn]*connManager)
		// broadcastMsgChan 用于广播消息
		broadcastMsgC := make(chan []byte, util.ChanCap)
		// updateHeartChan 更新心跳计数器
		updateHeartC := make(chan *net.TCPConn, util.ChanCap)
		// sendHeartChan 通知向客户端发送心跳包
		sendHeartC := make(chan *net.TCPConn, util.ChanCap)
		// connCloseChan 当读, 写线程出错, 自行选择关闭
		connCloseC := make(chan *net.TCPConn, util.ChanCap)

		ticker := time.NewTicker(util.TimeoutUnit)
		for {
			select {
			case conn := <-m.ConnJoinC:
				log.Printf("ConnJoinChan...")

				// 将connManager传入connMap
				sendMsgChan := make(chan []byte)
				writerStopC, readerStopC, msgSenderStopC, heartSenderStopC := make(chan struct{}, 1),
					make(chan struct{}, 1), make(chan struct{}, 1), make(chan struct{}, 1)
				connMap[conn] = &connManager{SendMsgChan: sendMsgChan, WriterStopC: writerStopC,
					ReaderStopC: readerStopC, MsgSenderStopC: msgSenderStopC, HeartSenderStopC: heartSenderStopC}

				// 开启读写协程
				writer.NewWriter(conn, sendMsgChan, connCloseC, writerStopC).Start()
				reader.NewReader(conn, broadcastMsgC, updateHeartC, sendHeartC, connCloseC, readerStopC).Start()

				// 欢迎消息
				welcomeMsg := (&util.Pack{Type: "message",
					Msg: fmt.Sprintf("欢迎[%v]进入聊天室", conn.RemoteAddr())}).Marshal()
				go func() {
					broadcastMsgC <- welcomeMsg
				}()
			case conn := <-updateHeartC:
				log.Printf("UpdateHeartChan...")
				cm, ok := connMap[conn]
				if ok {
					cm.TimeoutCount = 0
				}
			case msg := <-broadcastMsgC:
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
			case conn := <-sendHeartC:
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
			case conn := <-connCloseC:
				log.Printf("ConnCloseChan...")
				cm, ok := connMap[conn]
				if ok {
					remoteAddr := conn.RemoteAddr()
					_ = conn.Close()
					cm.MsgSenderStopC <- struct{}{}
					cm.HeartSenderStopC <- struct{}{}
					cm.WriterStopC <- struct{}{}
					cm.ReaderStopC <- struct{}{}
					delete(connMap, conn)
					byeMsg := (&util.Pack{Type: "message",
						Msg: fmt.Sprintf("用户[%v]离开了聊天室", remoteAddr)}).Marshal()
					go func() {
						broadcastMsgC <- byeMsg
					}()
				}
			case <-ticker.C:
				log.Printf("Ticker...")
				for conn, cm := range connMap {
					cm.TimeoutCount++
					if cm.TimeoutCount >= util.TimeoutMax {
						_ = conn.Close()
						cm.MsgSenderStopC <- struct{}{}
						cm.HeartSenderStopC <- struct{}{}
						cm.WriterStopC <- struct{}{}
						cm.ReaderStopC <- struct{}{}
						delete(connMap, conn)
					}
				}
			}
		}
	}()
}
