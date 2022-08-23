package main

// database module
// Copyright (c) 2022 - Valentin Kuznetsov <vkuznet@gmail.com>
//
// Go database APIs: http://go-database-sql.org/overview.html
// Oracle drivers:
//   _ "gopkg.in/rana/ora.v4"
//   _ "github.com/mattn/go-oci8"
// MySQL driver:
//   _ "github.com/go-sql-driver/mysql"
// SQLite driver:
//  _ "github.com/mattn/go-sqlite3"
//
// For full list of supported databases in Go please refer to
// https://github.com/golang/go/wiki/SQLDrivers
// All of them can be added in this package but so far we concentrate on traditional RDBMS

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gookit/color"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-oci8"
	_ "github.com/mattn/go-sqlite3"
)

// DB represents sql DB pointer
var DB *sql.DB

// DB transaction object
var TX *sql.Tx

// DBTYPE represents DBS DB type, e.g. ORACLE or SQLite
var DBTYPE string

// INDEX represents index value to use when printing DB records
// by default we use first record
var INDEX = 0

// LIMIT represents limit value to use when printing DB records
// by default there is no limit
var LIMIT = 10

// Record represents DBS record
type Record map[string]interface{}

// DBSQL represents DBS SQL record
var DBSQL Record

// DBOWNER represents DBS DB owner
var DBOWNER string

// helper function to initialize DB access
func dbInit(dburi string) (*sql.DB, error) {
	if strings.HasPrefix(dburi, "sqlite") {
		DBTYPE = "sqlite3"
	} else if strings.HasPrefix(dburi, "mysql") {
		DBTYPE = "mysql"
	} else if strings.HasPrefix(dburi, "postgres") {
		DBTYPE = "postgres"
	} else {
		DBTYPE = "oci8"
	}
	db, dberr := sql.Open(DBTYPE, parseDBUri(DBTYPE, dburi))
	if dberr != nil {
		log.Printf("unable to open %s, error %v", DBTYPE, dburi)
		return nil, dberr
	}
	dberr = db.Ping()
	if dberr != nil {
		log.Println("DB ping error", dberr)
		return nil, dberr
	}
	return db, nil
}

// cleanStatement cleans the given SQL statement to remove empty strings, etc.
func cleanStatement(stm string) string {
	var out []string
	for _, s := range strings.Split(stm, "\n") {
		//         s = strings.Trim(s, " ")
		if s == "" || s == " " {
			continue
		}
		out = append(out, s)
	}
	stm = strings.Join(out, "\n")
	return stm
}

// helper function to execute different SQL statements
func executeSQL(stm string, args ...interface{}) error {
	var err error
	if strings.HasPrefix(strings.ToLower(stm), "begin") {
		TX, err = DB.Begin()
		if err != nil {
			log.Println(err)
			return errors.New("unable to start transaction")
		}
	} else if strings.HasPrefix(strings.ToLower(stm), "commit") {
		if TX != nil {
			err = TX.Commit()
			if err != nil {
				log.Println(err)
				return errors.New("unable to commit transaction")
			}
		} else {
			return errors.New("Transaction was not started yet")
		}
	} else if strings.HasPrefix(strings.ToLower(stm), "rollback") {
		if TX != nil {
			err = TX.Rollback()
			if err != nil {
				log.Println(err)
				return errors.New("unable to rollback transaction")
			}
		}
	} else if strings.HasPrefix(strings.ToLower(stm), "insert") ||
		strings.HasPrefix(strings.ToLower(stm), "delete") {
		if TX != nil {
			_, err = TX.Exec(stm, args...)
			if err != nil {
				log.Println(stm, "error", err)
				return errors.New("unable to execute statement")
			}
		}
	} else {
		err = execute(stm, args...)
		if err != nil {
			log.Println("db error:", err)
			return err
		}
	}
	return err
}

// generic API to execute given statement
// ideas are taken from
// http://stackoverflow.com/questions/17845619/how-to-call-the-scan-variadic-function-in-golang-using-reflection
//gocyclo:ignore
func execute(stm string, args ...interface{}) error {
	stm = cleanStatement(stm)

	// execute transaction
	tx, err := DB.Begin()
	if err != nil {
		return Error(err, TransactionErrorCode, "", "execute")
	}
	defer tx.Rollback()
	rows, err := tx.Query(stm, args...)
	if err != nil {
		msg := fmt.Sprintf("unable to query statement: %v", stm)
		fmt.Println()
		log.Println(msg)
		return Error(err, QueryErrorCode, "", "execute")
	}
	defer rows.Close()

	// initialize tabwriter
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, MinWidth, TabWidth, Padding, ' ', 0)
	defer w.Flush()

	// extract columns from Rows object and create values & valuesPtrs to retrieve results
	columns, _ := rows.Columns()
	var cols []string
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	rowCount := 0
	if DBFORMAT == "rows" {
		fmt.Println("")
	}
	for rows.Next() {
		if rowCount == 0 {
			// initialize value pointers
			for i := range columns {
				valuePtrs[i] = &values[i]
			}
		}
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return Error(err, RowsScanErrorCode, "", "execute")
		}
		// store results into generic record (a dict)
		rec := make(Record)
		for i, col := range columns {
			if len(cols) != len(columns) {
				cols = append(cols, strings.ToLower(col))
			}
			vvv := values[i]
			switch val := vvv.(type) {
			case *sql.NullString:
				v, e := val.Value()
				if e == nil {
					rec[cols[i]] = v
				}
			case *sql.NullInt64:
				v, e := val.Value()
				if e == nil {
					rec[cols[i]] = v
				}
			case *sql.NullFloat64:
				v, e := val.Value()
				if e == nil {
					rec[cols[i]] = v
				}
			case *sql.NullBool:
				v, e := val.Value()
				if e == nil {
					rec[cols[i]] = v
				}
			default:
				rec[cols[i]] = val
			}
		}
		printRecord(w, rec, rowCount)
		rowCount += 1
	}
	// add additional print at the end
	if DBFORMAT == "rows" {
		fmt.Fprintf(w, "\n")
	} else {
		fmt.Println()
	}
	if err = rows.Err(); err != nil {
		return Error(err, RowsScanErrorCode, "", "execute")
	}
	return nil
}

// helper function to print DB record
func printRecord(w io.Writer, rec Record, rowCount int) {
	// do not print if we are out of range
	if rowCount < INDEX || (LIMIT > 0 && rowCount > LIMIT) {
		return
	}
	var maxKeyLength int
	var keys []string
	for key := range rec {
		if len(key) > maxKeyLength {
			maxKeyLength = len(key)
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if DBFORMAT == "pairs" {
		fmt.Println("")
		for _, key := range keys {
			val, _ := rec[key]
			var pad string
			for i := 0; i < maxKeyLength-len(key); i++ {
				pad += " "
			}
			if COLOR {
				fmt.Printf("%s%s: %v\n", color.Notice.Sprintf(key), pad, val)
			} else {
				fmt.Printf("%s%s: %v\n", key, pad, val)
			}
		}
		return
	} else if DBFORMAT == "json" {
		data, err := json.Marshal(rec)
		if err == nil {
			fmt.Println(string(data))
		}
		return
	}

	// print column names if necessary
	if rowCount == 0 {
		fmt.Fprintf(w, strings.Join(keys, "\t"))
		fmt.Fprintf(w, "\n")
	}

	// print row values
	var vals []string
	for _, key := range keys {
		val, _ := rec[key]
		vals = append(vals, fmt.Sprintf("%v", val))
	}
	fmt.Fprintf(w, strings.Join(vals, "\t")+"\n")
}
