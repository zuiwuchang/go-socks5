package main

import (
	"bytes"
	"fmt"
	"strings"
)

//配置 檔案
type Configure struct {
	//服務監聽 地址
	LAddr string `json:"LAddr"`
	//驗證密碼 如果為空 則不進行驗證
	Pwd string

	//要顯示的日誌 信息 all == trace,debug,info,warn,error,fault
	Logs string

	//recv 緩衝區 大小
	RecvBuffer int

	//是否使用 未加密的 h2c 模式
	H2C bool `json:"H2C"`

	//僅在使用 h2 時 指定 https 證書 路徑
	Crt string
	//僅在使用 h2 時 指定 https 證書 key 路徑
	Key string

	//客戶端 證書
	ClientCrts []string
}

func (c *Configure) String() string {
	w := bytes.NewBufferString("{\n")
	w.WriteString(fmt.Sprintf("	LAddr = %v,\n", c.LAddr))
	w.WriteString(fmt.Sprintf("	Pwd = %v,\n", c.Pwd))
	w.WriteString(fmt.Sprintf("	Logs = %v,\n", c.Logs))
	w.WriteString(fmt.Sprintf("	RecvBuffer = %v,\n", c.RecvBuffer))
	w.WriteString(fmt.Sprintf("	H2C = %v,\n", c.H2C))
	w.WriteString(fmt.Sprintf("	Crt = %v,\n", c.Crt))
	w.WriteString(fmt.Sprintf("	Key = %v,\n", c.Key))
	w.WriteString(fmt.Sprintf("	ClientCrts = %v,\n", c.ClientCrts))
	w.WriteString("}")
	return w.String()
}

func (c *Configure) Format() {
	c.LAddr = strings.TrimSpace(c.LAddr)
	if c.LAddr == "" {
		c.LAddr = ":2921"
	}

	c.Logs = strings.TrimSpace(c.Logs)

	if c.RecvBuffer < 1024 {
		c.RecvBuffer = 1024 * 32
	}

	if c.Crt == "" {
		c.Crt = DefaultCrt
	}
	if c.Key == "" {
		c.Key = DefaultKey
	}
}
