package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
}

func (c *Configure) String() string {
	w := bytes.NewBufferString("{\n")
	w.WriteString(fmt.Sprintf("	LAddr = %v,\n", c.LAddr))
	w.WriteString(fmt.Sprintf("	RemoteAddr = %v,\n", c.RemoteAddr))
	w.WriteString(fmt.Sprintf("	RemotePwd = %v,\n", c.RemotePwd))
	w.WriteString(fmt.Sprintf("	Logs = %v,\n", c.Logs))
	w.WriteString(fmt.Sprintf("	Timeout = %v,\n", c.Timeout))
	w.WriteString(fmt.Sprintf("	RecvBuffer = %v,\n", c.RecvBuffer))
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

	if c.Timeout == 0 {
		c.Timeout = time.Second * 15
	} else {
		c.Timeout *= time.Second
	}

	if c.RecvBuffer < 1024 {
		c.RecvBuffer = 1024 * 32
	}
}

func LoadConfigure(filename string) (cnf *Configure, e error) {
	var b []byte
	if b, e = ioutil.ReadFile(filename); e != nil {
		return
	}
	var configure Configure
	if e = json.Unmarshal(b, &configure); e != nil {
		return
	}
	cnf = &configure
	return
}
