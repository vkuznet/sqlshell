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
- persiste history
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
>

# any unix command is supported, e.g.
> ls
file1 file2

# use SQL command, by default it will use pairs format which shows
# given DB record as key:value pair printed separately
> select * from table;

id   : 1
name : value1

id   : 2
name : value2
...

# change database format
> dbformat
dbformat: json,pairs,rows or rows:minwidth:tabwidth:padding:padchar
Example : dbformat=rows:4:16:0

# setup db output as rows data-format
> dbformat=rows

# execute query
> select * from table;
1 value1
2 value2

# show history
> history
ls
pwd
select * from table;

# execute certain command from the history
> !3

id   : 1
name : value1

id   : 2
name : value2

```
