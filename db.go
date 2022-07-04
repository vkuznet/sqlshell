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
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-oci8"
	_ "github.com/mattn/go-sqlite3"
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
func dbInit(dburi string) (*sql.DB, error) {
	if strings.HasPrefix(dburi, "sqlite") {
		DBTYPE = "sqlite3"
		dburi = strings.Replace(dburi, "sqlite://", "", -1)
		dburi = strings.Replace(dburi, "sqlite3://", "", -1)
	} else if strings.HasPrefix(dburi, "mysql") {
		DBTYPE = "mysql"
	} else if strings.HasPrefix(dburi, "postgres") {
		DBTYPE = "postgres"
	} else {
		DBTYPE = "oci8"
	}
	db, dberr := sql.Open(DBTYPE, dburi)
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

// generic API to execute given statement
// ideas are taken from
// http://stackoverflow.com/questions/17845619/how-to-call-the-scan-variadic-function-in-golang-using-reflection
//gocyclo:ignore
func execute(stm string, args ...any) error {
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

	// extract columns from Rows object and create values & valuesPtrs to retrieve results
	columns, _ := rows.Columns()
	var cols []string
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	rowCount := 0
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
		printRecord(rec, rowCount)
		rowCount += 1
	}
	if err = rows.Err(); err != nil {
		return Error(err, RowsScanErrorCode, "", "execute")
	}
	return nil
}

// helper function to print DB record
func printRecord(rec Record, rowCount int) {
	var maxKeyLength int
	var keys []string
	for key, _ := range rec {
		if len(key) > maxKeyLength {
			maxKeyLength = len(key)
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	fmt.Println("")
	if DBFORMAT == "" || DBFORMAT == "cols" {
		for _, key := range keys {
			val, _ := rec[key]
			var pad string
			for i := 0; i < maxKeyLength-len(key); i++ {
				pad += " "
			}
			fmt.Printf("%s%s: %v\n", key, pad, val)
		}
		return
	} else if DBFORMAT == "json" {
		data, err := json.Marshal(rec)
		if err == nil {
			fmt.Println(string(data))
		}
		return
	}
	// initialize tabwriter
	w := new(tabwriter.Writer)
	// minwidth, tabwidth, padding, padchar, flags
	w.Init(os.Stdout, 4, 16, 0, '\t', 0)
	defer w.Flush()

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
	fmt.Fprintf(w, strings.Join(vals, "\t"))
}
