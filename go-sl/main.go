package main

import (
	"flag"
	"os"
	"strings"
	"time"
)

func exit() {
	os.Exit(1)
}
func main() {
	var h, h2c, skip bool
	var f, l, p, r, logs, crt, key string
	var timeout int64
	var buffer int

	flag.BoolVar(&h, "h", false, "show help")

	flag.StringVar(&f, "f", "", "configure file path")
	flag.StringVar(&l, "l", "", "local socks5 listen addr (default localhost:1911 [memories 1911/10/10])")
	flag.StringVar(&p, "p", "", "remote auth password")
	flag.StringVar(&r, "r", "", "remote address")
	flag.StringVar(&logs, "logs", "",
		`set show logs
	[-logs all]			show all logs
	[-logs warn,error,fault]	show warn,error,fault logs [trace debug info]`)
	flag.Int64Var(&timeout, "timeout", 0, "create socks5 channel timeout (default 15s)")
	flag.IntVar(&buffer, "buffer", 0, "recv buffer (default 32k)")

	flag.BoolVar(&h2c, "h2c", false, "use http2 h2c not tls")
	flag.StringVar(&crt, "crt", "", "certificate file path of h2")
	flag.StringVar(&key, "key", "", "certificate key file path of h2")
	flag.BoolVar(&skip, "skip", false, "skip verify credentials")

	flag.Parse()
	if h {
		flag.PrintDefaults()
		return
	}

	//load cnf
	cnf := &Configure{}

	//format cnf
	l = strings.TrimSpace(l)
	if l != "" {
		cnf.LAddr = l
	}
	if p != "" {
		cnf.RemotePwd = p
	}
	r = strings.TrimSpace(r)
	if r != "" {
		cnf.RemoteAddr = r
	}
	logs = strings.TrimSpace(logs)
	if logs != "" {
		cnf.Logs = logs
	}
	if timeout > 0 {
		cnf.Timeout = time.Duration(timeout) * time.Second
	}
	cnf.RecvBuffer = buffer
	cnf.H2C = h2c
	crt = strings.TrimSpace(crt)
	if crt != "" {
		cnf.Crt = crt
	}
	key = strings.TrimSpace(key)
	if key != "" {
		cnf.Key = key
	}
	cnf.SkipVerify = skip

	cnf.Format()
	initLogs(cnf)
	if Trace != nil {
		Trace.Printf("\n%v\n", cnf)
	}

	//run service
	var s Service
	s.runService(cnf)
}
