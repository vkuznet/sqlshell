package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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
		pgUri := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", c.Host, c.Port, c.User, c.Password, c.DBName)
		s = fmt.Sprintf("postgres://%s", pgUri)
	} else {
		log.Fatal(fmt.Sprintf("unsupported DB type %s", c.DBType))
	}
	return s
}

const (
	sqliteConfig string = `{"db_type": "sqlite", "db_path": "/tmp/file.db"}`
	oracleConfig string = `{"db_type": "oracle", "user":"bla", "password":"xyz", "dbname":"db"}`
	mysqlConfig  string = `{"db_type": "mysql", "user":"bla", "password":"xyz", "dbname":"db", "host": "127.0.0.1", "port":3306}`
	pgConfig     string = `{"db_type": "postgress", "user":"bla", "password":"xyz", "dbname":"db", "host":"127.0.0.1", "port":5432}`
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
