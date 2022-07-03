package main

import (
	"errors"
	"fmt"
	"strings"
)

// GenericErr represents generic dbs error
var GenericErr = errors.New("dbs error")

// DatabaseErr represents generic database error
var DatabaseErr = errors.New("database error")

// InvalidParamErr represents generic error for invalid input parameter
var InvalidParamErr = errors.New("invalid parameter(s)")

// NotImplementedApiErr represents generic not implemented api error
var NotImplementedApiErr = errors.New("not implemented api error")

// DBS Error codes provides static representation of DBS errors, they cover 1xx range
const (
	GenericErrorCode      = iota + 100 // generic DBS error
	DatabaseErrorCode                  // 101 database error
	TransactionErrorCode               // 102 transaction error
	QueryErrorCode                     // 103 query error
	RowsScanErrorCode                  // 104 row scan error
	SessionErrorCode                   // 105 db session error
	CommitErrorCode                    // 106 db commit error
	ParseErrorCode                     // 107 parser error
	GetIDErrorCode                     // 109 get id db error
	InsertErrorCode                    // 110 db insert error
	UpdateErrorCode                    // 111 update error
	LastInsertErrorCode                // 112 db last insert error
	ValidateErrorCode                  // 113 validation error
	DecodeErrorCode                    // 115 decode error
	EncodeErrorCode                    // 116 encode error
	NotImplementedApiCode              // 119 not implemented API error
	ReaderErrorCode                    // 120 io reader error
	WriterErrorCode                    // 121 io writer error
	UnmarshalErrorCode                 // 122 json unmarshal error
	MarshalErrorCode                   // 123 marshal error
)

// DBError represents common structure for DB errors
type DBError struct {
	Reason   string `json:"reason"`   // error string
	Message  string `json:"message"`  // additional message describing the issue
	Function string `json:"function"` // DB function
	Code     int    `json:"code"`     // DB error code
}

// Error function implements details of DBS error message
func (e *DBError) Error() string {
	sep := ": "
	if strings.Contains(e.Reason, "DBError") { // nested error
		sep += "nested "
	}
	return fmt.Sprintf(
		"DBError Code:%d Description:%s Function:%s Message:%s Error%s%v",
		e.Code, e.Explain(), e.Function, e.Message, sep, e.Reason)
}

func (e *DBError) Explain() string {
	switch e.Code {
	case GenericErrorCode:
		return "Generic DBS error"
	case DatabaseErrorCode:
		return "DBS DB error"
	case TransactionErrorCode:
		return "DBS DB transaction error"
	case QueryErrorCode:
		return "DBS DB query error, e.g. mailformed SQL statement"
	case RowsScanErrorCode:
		return "DBS DB row scane error, e.g. fail to get DB record from a database"
	case SessionErrorCode:
		return "DBS DB session error"
	case CommitErrorCode:
		return "DBS DB transaction commit error"
	case ParseErrorCode:
		return "DBS parser error, e.g. mailformed input parameter to the query"
	case GetIDErrorCode:
		return "DBS DB ID error for provided entity, e.g. there is no record in DB for provided value"
	case InsertErrorCode:
		return "DBS DB insert record error"
	case UpdateErrorCode:
		return "DBS DB update record error"
	case LastInsertErrorCode:
		return "DBS DB laster insert record error, e.g. fail to obtain last inserted ID"
	case ValidateErrorCode:
		return "DBS validation error, e.g. input parameter does not match lexicon rules"
	case DecodeErrorCode:
		return "DBS decode record failure, e.g. mailformed JSON"
	case EncodeErrorCode:
		return "DBS encode record failure, e.g. unable to convert structure to JSON"
	case NotImplementedApiCode:
		return "DBS Not implemented API error"
	case ReaderErrorCode:
		return "DBS reader I/O error, e.g. unable to read HTTP POST payload"
	case WriterErrorCode:
		return "DBS writer I/O error, e.g. unable to write record to HTTP response"
	case UnmarshalErrorCode:
		return "DBS unable to parse JSON record"
	case MarshalErrorCode:
		return "DBS unable to convert record to JSON"
	default:
		return "Not defined"
	}
	return "Not defined"
}

// helper function to create dbs error
func Error(err error, code int, msg, function string) error {
	reason := "nil"
	if err != nil {
		reason = err.Error()
	}
	return &DBError{
		Reason:   reason,
		Message:  msg,
		Code:     code,
		Function: function,
	}
}
