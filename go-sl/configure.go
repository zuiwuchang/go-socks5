package main

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

//配置 檔案
type Configure struct {
	//本地 socks5 監聽地址
	LAddr string `json:"LAddr"`

	//遠程 服務器 地址
	RemoteAddr string

	//遠程 服務器 驗證密碼
	RemotePwd string

	//要顯示的日誌 信息 all == trace,debug,info,warn,error,fault
	Logs string

	//建立 隧道 超時 時間 單位 秒
	Timeout time.Duration
	//recv 緩衝區 大小
	RecvBuffer int

	//是否使用 未加密的 h2c 模式
	H2C bool `json:"H2C"`
	//僅在使用 h2 時 指定 https 證書 路徑
	Crt string
	//僅在使用 h2 時 指定 https 證書 key 路徑
	Key string
	//不驗證 tls 證書
	SkipVerify bool
}

func (c *Configure) String() string {
	w := bytes.NewBufferString("{\n")
	w.WriteString(fmt.Sprintf("	LAddr = %v,\n", c.LAddr))
	w.WriteString(fmt.Sprintf("	RemoteAddr = %v,\n", c.RemoteAddr))
	w.WriteString(fmt.Sprintf("	RemotePwd = %v,\n", c.RemotePwd))
	w.WriteString(fmt.Sprintf("	Logs = %v,\n", c.Logs))
	w.WriteString(fmt.Sprintf("	Timeout = %v,\n", c.Timeout))
	w.WriteString(fmt.Sprintf("	RecvBuffer = %v,\n", c.RecvBuffer))
	w.WriteString(fmt.Sprintf("	H2C = %v,\n", c.H2C))
	w.WriteString(fmt.Sprintf("	Crt = %v,\n", c.Crt))
	w.WriteString(fmt.Sprintf("	Key = %v,\n", c.Key))
	w.WriteString(fmt.Sprintf("	SkipVerify = %v,\n", c.SkipVerify))
	w.WriteString("}")
	return w.String()
}
func (c *Configure) Format() {
	c.LAddr = strings.TrimSpace(c.LAddr)
	if c.LAddr == "" {
		c.LAddr = "localhost:1911"
	}
	c.RemoteAddr = strings.TrimSpace(c.RemoteAddr)

	c.Logs = strings.TrimSpace(c.Logs)

	if c.Timeout < 1 {
		c.Timeout = time.Second * 15
	}

	if c.RecvBuffer < 1024 {
		c.RecvBuffer = 1024 * 32
	}
}
