package main

import (
	klog "github.com/zuiwuchang/king-go/log"
	"log"
	"strings"
)

var Trace *log.Logger
var Debug *log.Logger
var Info *log.Logger
var Error *log.Logger
var Warn *log.Logger
var Fault *log.Logger
var g_logs *klog.Loggers = klog.NewDebugLoggers()

func init() {
	flags := log.Ltime
	//flags := log.Lshortfile | log.Lshortfile
	g_logs.Trace.SetFlags(flags)
	g_logs.Debug.SetFlags(flags)
	g_logs.Info.SetFlags(flags)
	g_logs.Error.SetFlags(flags)
	g_logs.Warn.SetFlags(flags)
	g_logs.Fault.SetFlags(flags)
}
func initLogs(cnf *Configure) {
	keys := make(map[string]bool)
	strs := strings.Split(strings.ToLower(cnf.Logs), ",")
	for _, str := range strs {
		str = strings.TrimSpace(str)
		if str != "" {
			keys[str] = true
		}
	}
	if len(keys) == 0 {
		return
	} else if _, ok := keys["all"]; ok {
		Trace = g_logs.Trace
		Debug = g_logs.Debug
		Info = g_logs.Info
		Warn = g_logs.Warn
		Error = g_logs.Error
		Fault = g_logs.Fault
		return
	} else if _, ok = keys["trace"]; ok {
		Trace = g_logs.Trace
	} else if _, ok = keys["debug"]; ok {
		Debug = g_logs.Debug
	} else if _, ok = keys["info"]; ok {
		Info = g_logs.Info
	} else if _, ok = keys["warn"]; ok {
		Warn = g_logs.Warn
	} else if _, ok = keys["error"]; ok {
		Error = g_logs.Error
	} else if _, ok = keys["fault"]; ok {
		Fault = g_logs.Fault
	}
}
