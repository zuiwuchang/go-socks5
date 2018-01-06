package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
}

func (c *Configure) String() string {
	w := bytes.NewBufferString("{\n")
	w.WriteString(fmt.Sprintf("	LAddr = %v,\n", c.LAddr))
	w.WriteString(fmt.Sprintf("	Pwd = %v,\n", c.Pwd))
	w.WriteString(fmt.Sprintf("	Logs = %v,\n", c.Logs))
	w.WriteString("}")
	return w.String()
}

func (c *Configure) Format() {
	c.LAddr = strings.TrimSpace(c.LAddr)
	if c.LAddr == "" {
		c.LAddr = ":2921"
	}

	c.Logs = strings.TrimSpace(c.Logs)
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
