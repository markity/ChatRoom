package util

import (
	"encoding/json"
	"net"
	"time"
)

// ServerAddr 运行地址
var ServerAddr = "127.0.0.1:8000"

// TimeoutUnit 超时时间
const TimeoutUnit = time.Second * 5

// TimeoutMax 最大超时次数, 达到则视为断开连接
const TimeoutMax = 3

// ChanCap 通道容量, 防止过多消息引起阻塞
const ChanCap = 128

// Pack 数据包
type Pack struct {
	Type string `json:"type"`
	Msg  string `json:"message"`
}

// Marshal 打包成bytes
func (p *Pack) Marshal() []byte {
	data, err := json.Marshal(p)
	// 除非Pack有问题, 否则不会失败
	if err != nil {
		panic(err)
	}
	return data
}

// Unmarshal 从bytes解包
func (p *Pack) Unmarshal(data []byte) error {
	err := json.Unmarshal(data, p)
	return err
}

// BoardcastMsgChan 用于广播消息
var BoardcastMsgChan = make(chan []byte, ChanCap)

// ConnJoinChan 通知有新的连接进入
var ConnJoinChan = make(chan *net.TCPConn, ChanCap)

// UpdateHeartChan 更新心跳计数器
var UpdateHeartChan = make(chan *net.TCPConn, ChanCap)

// SendHeartChan 通知向客户端发送心跳包
var SendHeartChan = make(chan *net.TCPConn, ChanCap)

// ConnCloseChan 当读, 写线程出错, 自行选择关闭
var ConnCloseChan = make(chan *net.TCPConn, ChanCap)
