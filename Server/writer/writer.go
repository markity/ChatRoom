package writer

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
)

func NewWriter(conn *net.TCPConn, sendMsgC chan []byte, connCloseC chan *net.TCPConn, stopC chan struct{}) *Writer {
	return &Writer{Conn: conn, SendMsgC: sendMsgC, ConnCloseC: connCloseC, StopC: stopC}
}

type Writer struct {
	Conn *net.TCPConn

	SendMsgC   chan []byte
	ConnCloseC chan *net.TCPConn
	StopC      chan struct{}
}

func (w *Writer) Start() {
	go func() {
		log.Printf("ConnWriter is running...")
		uint32Bytes := make([]byte, 4)
		for {
			select {
			case <-w.StopC:
				return
			case msg := <-w.SendMsgC:
				buf := bytes.NewBuffer(nil)
				binary.LittleEndian.PutUint32(uint32Bytes, uint32(len(msg)))
				_, err := buf.Write(uint32Bytes)
				if err != nil {
					w.ConnCloseC <- w.Conn
					return
				}
				_, err = buf.Write(msg)
				if err != nil {
					w.ConnCloseC <- w.Conn
					return
				}

				_, err = w.Conn.Write(buf.Bytes())
				if err != nil {
					w.ConnCloseC <- w.Conn
					return
				}
			}
		}
	}()
}
