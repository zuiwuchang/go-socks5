package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	kio "github.com/zuiwuchang/king-go/io"
	kstrings "github.com/zuiwuchang/king-go/strings"
	"go-socks5/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"time"
)

var errorCreateSocks5Timeout = errors.New("create socks5 timeout")
var errorBadSocks5Protocol = errors.New("unknow protocol")

type Service struct {
	Configure *Configure
	Client    pb.Socks5Client
}

func (s *Service) runService(cnf *Configure) {
	s.Configure = cnf
	//連接 服務器
	conn, e := grpc.Dial(cnf.RemoteAddr, grpc.WithInsecure())
	if e != nil {
		if Fault != nil {
			Fault.Fatalln(e)
		}
		exit()
	}
	defer conn.Close()
	s.Client = pb.NewSocks5Client(conn)

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
	//建立 socks5 超時
	t := time.AfterFunc(s.Configure.Timeout, func() {
		ch <- errorCreateSocks5Timeout
	})
	//建立 socks5
	var domain string
	var rs []byte
	var stream pb.Socks5_MakeChannelClient
	go func() {
		var e error
		rs, domain, e = s.createSocks5(c)
		if e != nil {
			ch <- e
			return
		}
		//建立 代理隧道
		stream, e = s.createChannel(c, domain, rs)
		if e != nil {
			ch <- e
			return
		}
		ch <- nil
	}()

	//等待 建立 socks5
	e := <-ch
	if e == nil {
		t.Stop()

		if Info != nil {
			Info.Println(c.RemoteAddr(), "->", domain)
		}
		s.connectChannel(c, stream)
		c.Close()
		return
	}
	c.Close()
	if e == errorCreateSocks5Timeout {
		if Warn != nil {
			Warn.Println(c.RemoteAddr(), "->", domain, e)
		}
	} else {
		t.Stop()
	}

}
func (s *Service) createSocks5(c net.Conn) (rs []byte, domain string, e error) {
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
	var n byte
	n, domain, e = s.getDomain(b[3], b[4:], c)
	if e != nil {
		return
	}
	b = b[:n+4]
	//驗證 命令
	if b[1] != 0x01 {
		//不支持的 命令
		//b[3] = 0x1
		b[1] = 0x07
		e = kio.WriteAll(c, b)
		if Error != nil {
			Error.Println(e)
		}
		return
	}
	rs = b
	return
}
func (s *Service) getDomain(artp byte, b []byte, c net.Conn) (n byte, str string, e error) {
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
		n = 6
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
		n = 1 + b[0] + 2
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
		n = 18
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
func (s *Service) createChannel(c net.Conn, domain string, rs []byte) (stream pb.Socks5_MakeChannelClient, e error) {
	if Debug != nil {
		Debug.Println("request", c.RemoteAddr(), "->", domain)
	}
	defer func() {
		if e != nil && stream != nil {
			stream.CloseSend()
		}
	}()

	stream, e = s.Client.MakeChannel(context.Background())
	if e != nil {
		if Warn != nil {
			Warn.Println(c.RemoteAddr(), "->", domain, e)
		}
		return nil, e
	}

	//通知 建立 連接
	var b []byte
	b, e = proto.Marshal(&pb.Connect{
		Addr: domain,
		Pwd:  s.Configure.RemotePwd,
	})
	if e != nil {
		if Error != nil {
			Error.Println(c.RemoteAddr(), "->", domain, e)
		}
		return
	}
	e = stream.SendMsg(&pb.Channel{
		Data: b,
	})
	if e != nil {
		if Warn != nil {
			Warn.Println(c.RemoteAddr(), "->", domain, e)
		}
		return
	}
	var connectRs pb.ConnectRs
	var m pb.Channel
	e = stream.RecvMsg(&m)
	if e != nil {
		if Warn != nil {
			Warn.Println(c.RemoteAddr(), "->", domain, e)
		}
		return
	}
	e = proto.Unmarshal(m.Data, &connectRs)
	if e != nil || connectRs.Code != 0 {
		if Warn != nil {
			Warn.Println(c.RemoteAddr(), "->", domain, e)
		}
		return
	}
	//通知 建立 隧道 成功
	rs[1] = 0
	e = kio.WriteAll(c, rs)
	if e != nil {
		if Error != nil {
			Error.Println(e)
			return
		}
	}
	return
}
func (s *Service) connectChannel(c net.Conn, stream pb.Socks5_MakeChannelClient) {
	ch := make(chan bool, 1)
	go s.forwardToRemote(ch, c, stream)
	go s.forwardFromRemote(ch, c, stream)
	<-ch
}
func (s *Service) forwardToRemote(ch chan bool, c net.Conn, stream pb.Socks5_MakeChannelClient) {
	b := make([]byte, s.Configure.RecvBuffer)
	var n int
	var e error
	var req pb.Channel
	for {
		n, e = c.Read(b)
		if e != nil {
			if Warn != nil {
				Warn.Println(e)
			}
			break
		}

		req.Data = b[:n]
		e = stream.SendMsg(&req)
		if e != nil {
			if Warn != nil {
				Warn.Println(e)
			}
			break
		}

	}
	ch <- true
}
func (s *Service) forwardFromRemote(ch chan bool, c net.Conn, stream pb.Socks5_MakeChannelClient) {
	var repl pb.Channel
	var e error
	for {
		e = stream.RecvMsg(&repl)
		if e != nil {
			if Warn != nil {
				Warn.Println(e)
			}
			break
		}

		e = kio.WriteAll(c, repl.Data)
		if e != nil {
			if Warn != nil {
				Warn.Println(e)
			}
			break
		}
	}
	ch <- true
}
