package main

import (
	"flag"
	"os"
	"strings"
)

const (
	DefaultCrt = "go-ss.crt"
	DefaultKey = "go-ss.key"
)

func exit() {
	os.Exit(1)
}
func main() {
	var h, h2c bool
	var p, logs, crt, key, crts string
	var buffer int

	flag.BoolVar(&h, "h", false, "show help")

	flag.StringVar(&p, "p", "", "auth password")
	flag.StringVar(&logs, "logs", "",
		`set show logs
	[-logs all]			show all logs
	[-logs warn,error,fault]	show warn,error,fault logs [trace debug info]`)
	flag.IntVar(&buffer, "buffer", 0, "recv buffer (default 32k)")

	flag.BoolVar(&h2c, "h2c", false, "use http2 h2c not tls")
	flag.StringVar(&crt, "crt", "", "certificate file path of h2")
	flag.StringVar(&key, "key", "", "certificate key file path of h2")
	flag.StringVar(&crts, "crts", "", "client certificate file path of h2 [c0.ctr:c1.crt:...]")
	flag.Parse()

	if h {
		flag.PrintDefaults()
		return
	}

	//load cnf
	cnf := &Configure{}

	//format cnf
	if p != "" {
		cnf.Pwd = p
	}
	logs = strings.TrimSpace(logs)
	if logs != "" {
		cnf.Logs = logs
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
	crts = strings.TrimSpace(crts)
	if crts != "" {
		strs := strings.Split(crts, ":")
		cnf.ClientCrts = make([]string, 0, len(strs))
		for _, str := range strs {
			str = strings.TrimSpace(str)
			if str != "" {
				cnf.ClientCrts = append(cnf.ClientCrts, str)
			}
		}
	}
	cnf.Format()

	initLogs(cnf)
	if Trace != nil {
		Trace.Printf("\n%v\n", cnf)
	}
	//run service
	var s Service
	s.runService(cnf)
}
