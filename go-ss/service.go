package main

import (
	"crypto/tls"
	"crypto/x509"
	"go-socks5/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"net"
)

type Service struct {
}

func (s *Service) runService(cnf *Configure) {
	if cnf.H2C {
		s.runH2C(cnf)
	} else {
		s.runH2(cnf)
	}
}
func (s *Service) runH2(cnf *Configure) {
	var creds credentials.TransportCredentials
	var e error
	if len(cnf.ClientCrts) == 0 {
		creds, e = credentials.NewServerTLSFromFile(cnf.Crt, cnf.Key)
		if e != nil {
			if Fault != nil {
				Fault.Println(e)
			}
			exit()
		}
	} else {
		//加載 x509 證書
		var cert tls.Certificate
		cert, e = tls.LoadX509KeyPair(cnf.Crt, cnf.Key)
		if e != nil {
			if Fault != nil {
				Fault.Println(e)
			}
			exit()
		}

		pool := x509.NewCertPool()
		var pem []byte
		for _, filename := range cnf.ClientCrts {
			pem, e = ioutil.ReadFile(filename)
			if e != nil {
				if Fault != nil {
					Fault.Println(e)
				}
				exit()
			}
			ok := pool.AppendCertsFromPEM(pem)
			if !ok {
				if Fault != nil {
					Fault.Println("can't add pem to pool", filename)
				}
				exit()
			}
		}

		//tls
		tlsConfigure := &tls.Config{
			Certificates: []tls.Certificate{cert},

			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  pool,
		}

		creds = credentials.NewTLS(tlsConfigure)
	}

	//創建 監聽 Listener
	l, e := net.Listen("tcp", cnf.LAddr)
	if e != nil {
		if Fault != nil {
			Fault.Println(e)
		}
		exit()
	}
	g_logs.Info.Println("h2 work at", cnf.LAddr)

	//創建 rpc 服務器
	gs := grpc.NewServer(
		grpc.Creds(creds),
	)

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
func (s *Service) runH2C(cnf *Configure) {
	//創建 監聽 Listener
	l, e := net.Listen("tcp", cnf.LAddr)
	if e != nil {
		if Fault != nil {
			Fault.Println(e)
		}
		exit()
	}
	g_logs.Info.Println("h2c work at", cnf.LAddr)

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
