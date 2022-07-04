# sqlshell
SQL shell: a (better) replacement for DB shell(s).

### Introduction
Each database has its own DB CLI tool, e.g. `sqlite3` or `sqlplus`, and some are
better or worse than others. For instance, the `sqlplus` does not support
proper cursor movement (it is not based on ncurses), and therefore lack of
useful features such as history, in-place query editing, etc.

This code provides uniform shell for different database, and therefore
it work identically regardless of underlying DB. The DB access is provided
via Go [`sql` module](http://go-database-sql.org/) and can support any
database via its their drivers.

### Usage
The *sqlshell* provides the following set of features:
- full access to UNIX commands
- full access to SQL commands
  - different output formatting options, e.g. columns, rows or json views
- persiste history
- uniform access to different database backend
  - currently sqlshell supports access to SQLite, MySQL, ORACLE, Postgres
    databases

Here is a preview of sqlshell session:

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
