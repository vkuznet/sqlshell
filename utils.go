package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// DBConfif represents DB configuration file
type DBConfig struct {
	DBType   string `json:"type"`     // database type, e.g. oracle or sqlite
	User     string `json:"user"`     // databae user name
	Password string `json:"password"` // database password
	DBName   string `json:"name"`     // database name, e.g. oracle DB alias name
	DBFile   string `json:"file"`     // database path, e.g. sqlite db file name
	Host     string `json:"host"`     // database hostname
	Port     int    `json:"port"`     // database port
}

// DBUri provides generic URI as <dbType>://<dburi>
func (c *DBConfig) DBUri() string {
	var s string
	if c.DBType == "sqlite" || c.DBType == "sqlite3" {
		s = fmt.Sprintf("sqlite://%s", c.DBFile)
	} else if c.DBType == "mysql" {
		// user:password@/dbname
		s = fmt.Sprintf("mysql://%s/%s@tcp{%s:%d}/%s", c.User, c.Password, c.Host, c.Port, c.DBName)
	} else if c.DBType == "oracle" {
		// oracle://user:password@dbname
		s = fmt.Sprintf("oracle://%s/%s@%s", c.User, c.Password, c.DBName)
	} else if c.DBType == "postgres" {
		pgUri := fmt.Sprintf("host=%s port=%v user=%s password=%s dbname=%s sslmode=disable", c.Host, c.Port, c.User, c.Password, c.DBName)
		s = fmt.Sprintf("postgres://%s", pgUri)
	} else {
		log.Fatal(fmt.Sprintf("unsupported DB type %s", c.DBType))
	}
	return s
}

const (
	sqliteConfig string = `{"type": "sqlite", "path": "/tmp/file.db"}`
	oracleConfig string = `{"type": "oracle", "user":"bla", "password":"xyz", "name":"db"}`
	mysqlConfig  string = `{"type": "mysql", "user":"bla", "password":"xyz", "name":"db", "host": "127.0.0.1", "port":3306}`
	pgConfig     string = `{"type": "postgress", "user":"bla", "password":"xyz", "name":"db", "host":"127.0.0.1", "port":5432}`
)

// helper function to read DB config file and return dburi
func readConfig(dburi string) string {
	file, err := os.Open(dburi)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	data, err := os.ReadFile(file.Name())
	if err != nil {
		log.Fatal(err)
	}
	var config DBConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}
	return config.DBUri()
}

// helper function to parse DB URI according to specific GoLang DB driver
func parseDBUri(dbtype, dburi string) string {
	if dbtype == "sqlite" || dbtype == "sqlite3" {
		dburi = strings.Replace(dburi, "sqlite://", "", -1)
		dburi = strings.Replace(dburi, "sqlite3://", "", -1)
	} else if dbtype == "mysql" {
		// user:password@tcp(127.0.0.1:3306)/hello
		return dburi
	} else if dbtype == "oracle" {
		// user/password@dbname
		return dburi
	} else if dbtype == "postgres" {
		// user:password@dbname:host:port
		arr := strings.Split(dburi, "@")
		if len(arr) != 2 {
			log.Fatal("wrong dburi, should be in form postgres://user:password@dbname:host:port")
		}
		brr := strings.Split(arr[0], ":")
		if len(brr) != 2 {
			log.Fatal("wrong dburi, should be in form postgres://user:password@dbname:host:port")
		}
		user := brr[0]
		password := brr[1]
		brr = strings.Split(arr[1], ":")
		if len(brr) != 3 {
			log.Fatal("wrong dburi, should be in form postgres://user:password@dbname:host:port")
		}
		dbname := brr[0]
		host := brr[1]
		port := brr[2]
		dburi = fmt.Sprintf("host=%s port=%v user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	}
	return dburi
}
