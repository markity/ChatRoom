package writer

import (
	"bytes"
	"chatserver/util"
	"encoding/binary"
	"log"
	"net"
)

// ConnWriter 用于向conn写消息
func ConnWriter(conn *net.TCPConn, sendMsgChan <-chan []byte, stopC <-chan struct{}) {
	log.Printf("ConnWriter is running...")
	uint32Bytes := make([]byte, 4)
	for {
		select {
		case <-stopC:
			return
		case msg := <-sendMsgChan:
			buf := bytes.NewBuffer(nil)
			binary.LittleEndian.PutUint32(uint32Bytes, uint32(len(msg)))
			_, err := buf.Write(uint32Bytes)
			if err != nil {
				util.ConnCloseChan <- conn
				return
			}
			_, err = buf.Write(msg)
			if err != nil {
				util.ConnCloseChan <- conn
				return
			}

			_, err = conn.Write(buf.Bytes())
			if err != nil {
				util.ConnCloseChan <- conn
				return
			}
		}
	}
}
