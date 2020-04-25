package util

import (
	"encoding/json"
	"time"
)

// ServerAddr 运行地址
var ServerAddr = "127.0.0.1:8000"

// MsgMaxLen 一条消息的最大长度
const MsgMaxLen = 250

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
