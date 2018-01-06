package main

import (
	"github.com/golang/protobuf/proto"
	kio "github.com/zuiwuchang/king-go/io"
	"go-socks5/pb"
	"net"
)

type Socks5 struct {
	Configure *Configure
}

func (s *Socks5) MakeChannel(stream pb.Socks5_MakeChannelServer) (e error) {
	//獲取 連接 信息 地址
	var m pb.Channel
	e = stream.RecvMsg(&m)
	if e != nil {
		if Error != nil {
			Error.Println(e)
		}
		return
	}
	var connect pb.Connect
	e = proto.Unmarshal(m.Data, &connect)
	if e != nil {
		if Error != nil {
			Error.Println(e)
		}
		return
	} else if s.Configure.Pwd != "" && s.Configure.Pwd != connect.Pwd {
		if Warn != nil {
			Warn.Println("pwd not match")
		}
		return
	}

	var c net.Conn
	c, e = net.Dial("tcp", connect.Addr)
	if e != nil {
		if Error != nil {
			Error.Println(e)
		}

		//通知 建立 連接 失敗
		m.Data, e = proto.Marshal(&pb.ConnectRs{
			Code: 1,
			Emsg: e.Error(),
		})
		e = stream.SendMsg(&m)
		if e != nil && Error != nil {
			Error.Println(e)
		}
		return
	}
	defer c.Close()
	//通知 建立 連接 成功
	m.Data, e = proto.Marshal(&pb.ConnectRs{})
	e = stream.SendMsg(&m)
	if e != nil {
		if Error != nil {
			Error.Println(e)
		}
		return
	}

	ch := make(chan bool, 1)
	go func() {
		var e error
		var n int
		b := make([]byte, 1024*16)
		var relp pb.Channel
		for {
			n, e = c.Read(b)
			if e != nil {
				if Warn != nil {
					Warn.Println(e)
				}
				break
			}
			//fmt.Println("read ok")
			relp.Data = b[:n]
			e = stream.SendMsg(&relp)
			if e != nil {
				if Warn != nil {
					Warn.Println(e)
				}
				break
			}
			//fmt.Println("write ok")
		}
		ch <- true
	}()
	go func() {
		var req pb.Channel
		var e error
		for {
			e = stream.RecvMsg(&req)
			if e != nil {
				if Warn != nil {
					Warn.Println(e)
				}
				break
			}
			//fmt.Println("0 read ok")

			e = kio.WriteAll(c, req.Data)
			if e != nil {
				if Warn != nil {
					Warn.Println(e)
				}
				break
			}
			//fmt.Println("0 write ok")
		}
		ch <- true
	}()

	<-ch
	return
}
