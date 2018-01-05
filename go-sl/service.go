package main

import (
	"errors"
	kio "github.com/zuiwuchang/king-go/io"
	"net"
	"time"
)

var errorCreateSocks5Timeout = errors.New("create socks5 timeout")
var errorBadSocks5Protocol = errors.New("unknow protocol")

type Service struct {
	Configure *Configure
}

func (s *Service) runService(cnf *Configure) {
	s.Configure = cnf

	l, e := net.Listen("tcp", cnf.LAddr)
	if e != nil {
		if Fault != nil {
			Fault.Fatalln(e)
		}
		exit()
	}
	if Info != nil {
		Info.Println("socks5 work at", cnf.LAddr)
	}

	var c net.Conn
	for {
		c, e = l.Accept()
		if e == nil {
			go s.onConnect(c)
		} else {
			if Error != nil {
				Error.Println(e)
			}
		}
	}
}
func (s *Service) onConnect(c net.Conn) {
	if Trace != nil {
		Trace.Println("connect", c.RemoteAddr())
		defer Trace.Println("close", c.RemoteAddr())
	}
	ch := make(chan error, 1)
	//10秒內建立 socks5
	t := time.AfterFunc(time.Second*10, func() {
		ch <- errorCreateSocks5Timeout
	})
	//建立 socks5
	go func() {
		ch <- s.createSocks5(c)
	}()

	//等待 建立 socks5
	e := <-ch
	if e == nil {
		return
	}
	c.Close()
	if e != errorCreateSocks5Timeout {
		t.Stop()
	}
}
func (s *Service) createSocks5(c net.Conn) (e error) {
	//協商 驗證方式
	b := make([]byte, 7+255)
	e = kio.ReadAll(c, b[:2])
	if e != nil {
		if Error != nil {
			Error.Println(e)
		}
		return
	}
	//驗證 socks5 版本號
	if b[0] != 0x5 {
		e = errorBadSocks5Protocol
		if Error != nil {
			Error.Println(e)
		}
		return
	}
	if b[1] == 0 {
		e = errorBadSocks5Protocol
		if Error != nil {
			Error.Println(e)
		}
		return
	}
	e = kio.ReadAll(c, b[2:2+b[1]])
	if e != nil {
		if Error != nil {
			Error.Println(e)
		}
		return
	}
	yes := false
	for _, v := range b[2 : 2+b[1]] {
		if v == 0x00 {
			yes = true
			break
		}
	}
	if yes {
		//通知 客戶端 使用 0x00 驗證方式
		b[1] = 0x00
		e = kio.WriteAll(c, b[:2])
		if e != nil {
			if Error != nil {
				Error.Println(e)
			}
			return
		}
	} else {
		//通知 客戶端 沒有 合適的 認證方式
		b[1] = 0xff
		e = kio.WriteAll(c, b[:2])
		if e != nil {
			if Error != nil {
				Error.Println(e)
			}
			return
		}
		return
	}

	//等待 CONNECT 請求
	e = kio.ReadAll(c, b[:4])
	if e != nil {
		if Error != nil {
			Error.Println(e)
		}
		return
	}
	//驗證 socks5 版本號
	if b[0] != 0x5 {
		e = errorBadSocks5Protocol
		if Error != nil {
			Error.Println(e)
		}
		return
	}
	Info.Println(b)
	return
}
