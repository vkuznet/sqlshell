# sqlshell
`sqlshell` a (better) replacement of database (DB) shell(s)

[![Build](https://github.com/vkuznet/sqlshell/actions/workflows/build.yml/badge.svg)](https://github.com/vkuznet/sqlshell/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vkuznet/sqlshell)](https://goreportcard.com/report/github.com/vkuznet/sqlshell)
[![GoDoc](https://godoc.org/github.com/vkuznet/sqlshell?status.svg)](https://godoc.org/github.com/vkuznet/sqlshell)


### Introduction
Each database has its own Command Line (CLI) tool, e.g. `sqlite3` for SQLite
database, `sqlplus` for ORACLE one. All of these tools have
their own syntax and behavior and sometime lack of certain features
found in other tools. For instance, the `sqlplus` does not support
proper cursor movement (it is not based on ncurses), and therefore lack of
useful features such as history, in-place query editing, etc.

The `sqlshell` provides uniform shell for different database, and therefore
it works identically regardless of underlying DB. The DB access is provided
via Go [`sql` module](http://go-database-sql.org/) and can support any
database throught native database libraries.

### Build
To build `sqlshell` you'll need specific DB libraries, e.g.
- for SQLite you do not need anything to do
- for ORACLE please obtain ORACLE SDK and adjust accordingly oci8.pc file
- for MySQL ...
- for Postgress ...
After that, just run `make` to make a build on your architecture
or use `make build_linux`, etc. for other platforms

### Usage
The `sqlshell` provides the following set of features:
- full access to SQL commands
  - different output formatting options, e.g. columns, rows or json views
- persistent history
- uniform access to different database backend
  - currently sqlshell supports access to SQLite, MySQL, ORACLE, Postgres
    databases
- full access to UNIX commands, yes you can execute your favorite UNIX
command, e.g. ls or pwd or even vim :)

Here is a preview of `sqlshell` session:

```
# start from any UNIX shell
# sqlshell sqlite:///tmp/file.db

# now we are in sqlshell
sqlsh >

# any unix command is supported, e.g.
sqlsh > ls
file1 file2

# use SQL command, by default it will use pairs format which shows
# given DB record as key:value pair printed separately
sqlsh > select * from table;

id   : 1
name : value1

id   : 2
name : value2
...

# change database format
sqlsh > set format
format  : json,pairs,rows or rows:minwidth:tabwidth:padding:padchar
Example : dbformat=rows:4:16:0

# setup db output as rows data-format
sqlsh > set format=rows

# execute query
sqlsh > select * from table;
1 value1
2 value2

# show history
sqlsh > history
ls
pwd
select * from table;

# execute certain command from the history
sqlsh > !3

id   : 1
name : value1

id   : 2
name : value2

```
The `sqlshell` also adds useful `index,limit` commands to manage output from
database. Since different DBs use different methods (in MySQL you need to use
`LIMIT X, Y` while in ORALCE you need to wrap your SQL statement into another
one to use `ROWNU`) we provide this function to manage this use-cases. For
example, to limit your DB results to a specific range just do the following:
```
sqlsh > set index=10
sqlsh > set limit=100
sqlsh > select * from table;
```
and it will show only results within `index-limit` range (in this case
between 10 and 100).
