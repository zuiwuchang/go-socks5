package main

import (
	"flag"
	"os"
	"strings"
)

const (
	ConfigureFile = "go-ss.json"
)

func exit() {
	os.Exit(1)
}
func main() {
	var h bool
	var f, p, logs string

	flag.BoolVar(&h, "h", false, "show help")

	flag.StringVar(&f, "f", "", "configure file path")
	flag.StringVar(&p, "p", "", "auth password")
	flag.StringVar(&logs, "logs", "",
		`set show logs
	[-logs all]			show all logs
	[-logs warn,error,fault]	show warn,error,fault logs [trace debug info]`)
	flag.Parse()

	if h {
		flag.PrintDefaults()
		return
	}

	//load cnf
	var cnf *Configure
	if f == "" {
		cnf = &Configure{}
	} else {
		var e error
		cnf, e = LoadConfigure(f)
		if e != nil {
			g_logs.Fault.Fatalln(e)
			return
		}
	}
	//format cnf
	cnf.Format()
	if p != "" {
		cnf.Pwd = p
	}
	logs = strings.TrimSpace(logs)
	if logs != "" {
		cnf.Logs = logs
	}
	initLogs(cnf)
	if Trace != nil {
		Trace.Printf("\n%v\n", cnf)
	}
	//run service
	var s Service
	s.runService(cnf)
}
