package main

// sqlshell: a better replacement for DB shell(s)
// Copyright (c) 2015 - Valentin Kuznetsov <vkuznet@gmail.com>
//

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
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
	// initialize our DB connection
	dburi := strings.Join(os.Args[1:], "")
	db, dberr := dbInit(dburi)
	if dberr != nil {
		log.Fatal(dberr)
	}
	DB = db
	defer DB.Close()

	// set logger flags
	log.SetFlags(0)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// setup communication channels
	ch := make(chan string)
	cmdDone := make(chan bool)
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// run keyboard and command handlers to manage the shell
	go cmdHandler(ch, cmdDone)
	go keysHandler(ch)

	// shutdown our handlers
	<-done
	cmdDone <- true
}
