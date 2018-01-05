package main

import (
	"encoding/json"
	"io/ioutil"
)

//配置 檔案
type Configure struct {
	//本地 socks5 監聽地址
	LAddr string `json:"LAddr"`

	//遠程 服務器 地址
	RemoteAddr string
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
