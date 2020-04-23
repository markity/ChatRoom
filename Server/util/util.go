package util

import (
	"encoding/json"
	"net"
	"time"
)

// TimeoutUnit 超时时间
var TimeoutUnit = time.Second * 5

// TimeoutMax 最大超时次数, 达到则视为断开连接
var TimeoutMax = 3

// HeartPack 心跳包
var HeartPack = []byte(`{"type":"heart"}`)

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
var BoardcastMsgChan = make(chan []byte)

// ConnManager 存储关于线程的信息
type ConnManager struct {
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

// ConnMap 用于维护线程, 向线程发送消息
var ConnMap = make(map[*net.TCPConn]*ConnManager)

// ConnJoinChan 通知有新的连接进入
var ConnJoinChan = make(chan *net.TCPConn)

// UpdateHeartChan 更新心跳计数器
var UpdateHeartChan = make(chan *net.TCPConn)

// SendHeartChan 通知向客户端发送心跳包
var SendHeartChan = make(chan *net.TCPConn)

// ConnCloseChan 当读, 写线程出错, 自行选择关闭
var ConnCloseChan = make(chan *net.TCPConn)