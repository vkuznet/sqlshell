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
