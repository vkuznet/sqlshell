package main

// sqlshell: a better replacement for DB shell(s)
// Copyright (c) 2015 - Valentin Kuznetsov <vkuznet@gmail.com>
//

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

// version of the code
var gitVersion string

// Info function returns version string of the server
func info() string {
	goVersion := runtime.Version()
	tstamp := time.Now().Format("2006-02-01")
	return fmt.Sprintf("dbs2go git=%s go=%s date=%s", gitVersion, goVersion, tstamp)
}

func main() {
	ch := make(chan string)
	cmdDone := make(chan bool)
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go cmdHandler(ch, cmdDone)
	go keysHandler(ch)

	<-done
	cmdDone <- true
}
