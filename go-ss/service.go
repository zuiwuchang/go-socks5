package main

import (
	"go-socks5/pb"
	"google.golang.org/grpc"
	"net"
)

type Service struct {
}

func (s *Service) runService(cnf *Configure) {
	//創建 監聽 Listener
	l, e := net.Listen("tcp", cnf.LAddr)
	if e != nil {
		if Fault != nil {
			Fault.Println(e)
		}
		exit()
	}
	if Info != nil {
		Info.Println("work at", cnf.LAddr)
	}

	//創建 rpc 服務器
	gs := grpc.NewServer()

	//註冊 服務
	pb.RegisterSocks5Server(
		gs,
		&Socks5{
			Configure: cnf,
		},
	)

	//讓 rpc 在 Listener 上 工作
	if e = gs.Serve(l); e != nil {
		if Fault != nil {
			Fault.Fatalln(e)
		}
		exit()
	}
}
