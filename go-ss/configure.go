package main

import (
	"encoding/json"
	"io/ioutil"
)

//配置 檔案
type Configure struct {
	//服務監聽 地址
	LAddr string `json:"LAddr"`
	//驗證密碼 如果為空 則不進行驗證
	Pwd string
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
