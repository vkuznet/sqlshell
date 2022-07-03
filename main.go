package main

// sqlshell: a better replacement for DB shell(s)
// Copyright (c) 2015 - Valentin Kuznetsov <vkuznet@gmail.com>
//

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
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
	if len(os.Args) < 2 {
		fmt.Println("usage: sqlshell <dbtype://dburi>")
		fmt.Println("supported dbtypes: sqlite, mysql, postgres, oracle")
		os.Exit(1)
	}
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
	done := make(chan bool)

	// run keyboard and command handlers to manage the shell
	go cmdHandler(ch, done)
	keysHandler(ch)

	// shutdown our command handler
	done <- true
}
