package main

import (
	"flag"
	"os"
)

const (
	ConfigureFile = "go-sl.json"
)

func exit() {
	os.Exit(1)
}
func main() {
	var h bool
	var f string

	flag.BoolVar(&h, "h", false, "show help")

	flag.StringVar(&f, "f", ConfigureFile, "configure file path")

	flag.Parse()

	if h {
		flag.PrintDefaults()
		return
	}

	//load cnf
	cnf, e := LoadConfigure(f)
	if e != nil {
		if Fault != nil {
			Fault.Fatalln(e)
		}
		exit()
		return
	}
	//format cnf

	//run service
	var s Service
	s.runService(cnf)
}
