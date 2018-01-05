package main

import (
	klog "github.com/zuiwuchang/king-go/log"
	"log"
)

var Trace *log.Logger
var Debug *log.Logger
var Info *log.Logger
var Error *log.Logger
var Warn *log.Logger
var Fault *log.Logger

func init() {
	l := klog.NewDebugLoggers()

	//Trace = l.Trace
	Debug = l.Debug
	Info = l.Info
	Warn = l.Warn
	Error = l.Error
	Fault = l.Fault
}
