package main

// sqlshell: a better replacement for DB shell(s)
// Copyright (c) 2022 - Valentin Kuznetsov <vkuznet@gmail.com>
//

import (
	"errors"
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

// helper function to provide usage help string to stdout
func usage() {
	fmt.Println("Usage   : sqlshell <dbtype://dburi> or <dbConfigFile>")
	fmt.Println("DBTypes : sqlite, mysql, postgres, oracle")
	fmt.Println("Examples:")
	fmt.Println("          connect to SQLiteDB : sqlshell sqlite:///path/file.db")
	fmt.Println("          connect to ORACLE   : sqlshell oracle://user:password@dbname")
	fmt.Println("          connect to MySQL    : sqlshell mysql://user:password@/dbname")
	fmt.Println("          connect to Postgres : sqlshell postgres://user:password@dbname:host:port")
	fmt.Println("db configuration file examples:")
	fmt.Println("for SQLite  ", sqliteConfig)
	fmt.Println("for ORACLE  ", oracleConfig)
	fmt.Println("for MySQL   ", mysqlConfig)
	fmt.Println("for Postgres", pgConfig)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	// initialize our DB connection
	var dburi string
	arg := strings.Join(os.Args[1:], "")
	if _, err := os.Stat(arg); errors.Is(err, os.ErrNotExist) {
		dburi = arg
	} else {
		dburi = readConfig(arg)
	}

	// initialize access to DB
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

	//     reset()
	// shutdown our command handler
	done <- true

}
