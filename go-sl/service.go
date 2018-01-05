package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	kio "github.com/zuiwuchang/king-go/io"
	kstrings "github.com/zuiwuchang/king-go/strings"
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
	var domain string
	go func() {
		var e error
		domain, e = s.createSocks5(c)
		ch <- e
	}()

	//等待 建立 socks5
	e := <-ch
	if e == nil {
		//建立 代理隧道
		s.createChannel(c, domain)
		return
	}
	c.Close()
	if e != errorCreateSocks5Timeout {
		t.Stop()
	}
}
func (s *Service) createSocks5(c net.Conn) (domain string, e error) {
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
	domain, e = s.getDomain(b[3], b[4:], c)
	if e != nil {
		return
	}
	//驗證 命令
	if b[1] != 0x01 {
		//不支持的 命令
		b[3] = 0x1
		b[1] = 0x07
		e = kio.WriteAll(c, b[:10])
		if Error != nil {
			Error.Println(e)
		}
		return
	}

	return
}
func (s *Service) getDomain(artp byte, b []byte, c net.Conn) (str string, e error) {
	var domain string
	//獲取 地址 長度
	switch artp {
	case 0x1: //ipv4
		e = kio.ReadAll(c, b[:6])
		if e != nil {
			if Error != nil {
				Error.Println(e)
			}
			return
		}
		domain = fmt.Sprintf("%v:%v\n", net.IP(b[:4]).String(), binary.BigEndian.Uint16(b[4:]))
	case 0x3: //domain
		e = kio.ReadAll(c, b[:1])
		if e != nil {
			if Error != nil {
				Error.Println(e)
			}
			return
		}
		if b[0] == 0 {
			e = errorBadSocks5Protocol
			if Error != nil {
				Error.Println(e)
			}
			return
		}
		e = kio.ReadAll(c, b[1:1+b[0]+2])
		if e != nil {
			if Error != nil {
				Error.Println(e)
			}
			return
		}
		domain = fmt.Sprintf(
			"%v:%v",
			kstrings.BytesToString(b[1:1+b[0]]),
			binary.BigEndian.Uint16(b[1+b[0]:]),
		)
	case 0x4: //ipv6
		e = kio.ReadAll(c, b[:18])
		if e != nil {
			if Error != nil {
				Error.Println(e)
			}
			return
		}
		domain = fmt.Sprintf("%v:%v\n", net.IP(b[:16]).String(), binary.BigEndian.Uint16(b[16:]))
	default:
		e = errorBadSocks5Protocol
		if Error != nil {
			Error.Println(e)
		}
		return
	}

	str = domain
	return
}
func (s *Service) createChannel(c net.Conn, domain string) {
	defer c.Close()
	//Info.Println(domain)
	c0, e := net.Dial("tcp", domain)
	if e != nil {
		if Warn != nil {
			Warn.Println(e)
		}
		return
	}
	defer c0.Close()

	b := make([]byte, 1024*16)
	//通知 建立 隧道 成功
	b[0] = 0x5
	b[3] = 0x1
	e = kio.WriteAll(c, b[:10])
	if e != nil {
		if Error != nil {
			Error.Println(e)
			return
		}
	}
	ch := make(chan bool, 1)

	go func() {
		var n int
		for {
			n, e = c.Read(b)
			if e != nil {
				if Warn != nil {
					Warn.Println(e)
					break
				}
			}
			//fmt.Println("read ok")

			e = kio.WriteAll(c0, b[:n])
			if e != nil {
				if Warn != nil {
					Warn.Println(e)
					break
				}
			}
			//fmt.Println("write ok")
		}
		ch <- true
	}()
	go func() {
		b := make([]byte, 1024*16)
		var n int
		for {
			n, e = c0.Read(b)
			if e != nil {
				if Warn != nil {
					Warn.Println(e)
					break
				}
			}
			//fmt.Println("0 read ok")

			e = kio.WriteAll(c, b[:n])
			if e != nil {
				if Warn != nil {
					Warn.Println(e)
					break
				}
			}
			//fmt.Println("0 write ok")
		}
		ch <- true
	}()
	<-ch
}
