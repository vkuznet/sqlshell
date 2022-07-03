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

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
)

// DB represents sql DB pointer
var DB *sql.DB

// DBTYPE represents DBS DB type, e.g. ORACLE or SQLite
var DBTYPE string

// Record represents DBS record
type Record map[string]any

// DBSQL represents DBS SQL record
var DBSQL Record

// DBOWNER represents DBS DB owner
var DBOWNER string

// helper function to initialize DB access
func dbInit(dbtype, dburi string) (*sql.DB, error) {
	db, dberr := sql.Open(dbtype, dburi)
	if dberr != nil {
		log.Printf("unable to open %s, error %v", dbtype, dburi)
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

// generic API to execute given statement
// ideas are taken from
// http://stackoverflow.com/questions/17845619/how-to-call-the-scan-variadic-function-in-golang-using-reflection
//gocyclo:ignore
func execute(w io.Writer, sep, stm string, args ...interface{}) error {
	stm = cleanStatement(stm)

	var enc *json.Encoder
	if w != nil {
		enc = json.NewEncoder(w)
	}

	// execute transaction
	tx, err := DB.Begin()
	if err != nil {
		return Error(err, TransactionErrorCode, "", "dbs.executeAll")
	}
	defer tx.Rollback()
	rows, err := tx.Query(stm, args...)
	if err != nil {
		msg := fmt.Sprintf("unable to query statement: %v", stm)
		log.Println(msg)
		return Error(err, QueryErrorCode, "", "dbs.executeAll")
	}
	defer rows.Close()

	// extract columns from Rows object and create values & valuesPtrs to retrieve results
	columns, _ := rows.Columns()
	var cols []string
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	rowCount := 0
	writtenResults := false
	for rows.Next() {
		if rowCount == 0 {
			// initialize value pointers
			for i := range columns {
				valuePtrs[i] = &values[i]
			}
		}
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return Error(err, RowsScanErrorCode, "", "dbs.executeAll")
		}
		if rowCount != 0 && w != nil {
			// add separator line to our output
			w.Write([]byte(sep))
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
		if w != nil {
			if rowCount == 0 {
				if sep != "" {
					writtenResults = true
					w.Write([]byte("[\n"))
					defer w.Write([]byte("]\n"))
				}
			}
			err = enc.Encode(rec)
			if err != nil {
				return Error(err, EncodeErrorCode, "", "dbs.executeAll")
			}
		}
		rowCount += 1
	}
	if err = rows.Err(); err != nil {
		return Error(err, RowsScanErrorCode, "", "dbs.executeAll")
	}
	// make sure we write proper response if no result written
	if sep != "" && !writtenResults {
		w.Write([]byte("[]"))
	}
	return nil
}
