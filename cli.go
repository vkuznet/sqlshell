package main

// sqlshell: a better replacement for DB shell(s)
// Copyright (c) 2022 - Valentin Kuznetsov <vkuznet@gmail.com>
//
// the original idea is based on https://sj14.gitlab.io/post/2018/07-01-go-unix-shell/
// to fix "for sys error like go:linkname must refer to declared function or variable" do the following
// go get -u golang.org/x/sys
// https://stackoverflow.com/questions/71507321/go-1-18-build-error-on-mac-unix-syscall-darwin-1-13-go253-golinkname-mus
// ASCII codes: https://gist.github.com/fnky/458719343aabd01cfb17a3a4f7296797
// keyboard handling: https://github.com/atomicgo/keyboard
// cursor handling: https://github.com/atomicgo/cursor

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"atomicgo.dev/cursor"
	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/gookit/color"
)

// PROMPT represents shell prompt
var PROMPT = "sqlsh > "

// COLOR set color output
var COLOR bool

// HIST_LIMIT defines number of line we write to history file
var HIST_LIMIT = 100

// DBFORMAT defines how to print DB records, e.g. pairs or rows or json
// default is key:value pairs
var DBFORMAT string = "pairs"

// MinWidth used by tabwriter
var MinWidth int = 4

// TabWidth used by tabwriter
var TabWidth int = 4

// Padding used by tabwriter
var Padding int = 1

// helper function to handle keyboard input
//gocyclo:ignore
func keysHandler(ch chan<- string) {
	var pos, hpos int
	var cmd []string
	history := ReadHistory()
	if len(history) > 0 {
		hpos = len(history)
	}

	// start detecting user input command
	if COLOR {
		color.Info.Printf(PROMPT)
	} else {
		fmt.Printf(PROMPT)
	}
	err := keyboard.Listen(func(key keys.Key) (stop bool, err error) {

		switch key.Code {
		case keys.RuneKey:
			insert := false
			if len(cmd) > pos {
				cmd = append(cmd[:pos+1], cmd[pos:]...)
				cmd[pos] = key.String()
				insert = true
			} else {
				cmd = append(cmd, key.String())
			}
			pos += 1
			cursor.StartOfLine()
			if COLOR {
				fmt.Printf(color.Info.Sprintf(PROMPT) + strings.Join(cmd, ""))
			} else {
				fmt.Printf(PROMPT + strings.Join(cmd, ""))
			}
			if insert {
				cursor.Left(len(cmd) - pos)
			}
		case keys.Left:
			if pos > 0 {
				cursor.Left(1)
				pos -= 1
			}
		case keys.Right:
			if pos <= len(cmd) {
				cursor.Right(1)
				pos += 1
			}
		case keys.Up:
			if hpos > 0 {
				hpos -= 1
			}
			if len(history) > 0 && hpos < len(history) {
				cursor.StartOfLine()
				cursor.ClearLine()
				cmd = strings.Split(history[hpos], "")
				if COLOR {
					fmt.Printf(color.Info.Sprintf(PROMPT) + strings.Join(cmd, ""))
				} else {
					fmt.Printf(PROMPT + strings.Join(cmd, ""))
				}
				pos = len(cmd)
			}
		case keys.Down:
			if hpos < len(history) {
				hpos += 1
			}
			if len(history) > 0 && hpos < len(history) {
				cursor.StartOfLine()
				cursor.ClearLine()
				cmd = strings.Split(history[hpos], "")
				if COLOR {
					fmt.Printf(color.Info.Sprintf(PROMPT) + strings.Join(cmd, ""))
				} else {
					fmt.Printf(PROMPT + strings.Join(cmd, ""))
				}
				pos = len(cmd)
			}
		case keys.Space:
			cursor.Right(1)
			cmd = append(cmd, " ")
			pos += 1
		case keys.Backspace:
			if pos > 0 {
				front := cmd[:pos-1]
				var rest []string
				if pos < len(cmd) {
					rest = cmd[pos:]
				}
				pos -= 1
				cursor.StartOfLine()
				cursor.ClearLine()
				cmd = front
				cmd = append(cmd, rest...)
				if COLOR {
					fmt.Printf(color.Info.Sprintf(PROMPT) + strings.Join(cmd, ""))
				} else {
					fmt.Printf(PROMPT + strings.Join(cmd, ""))
				}
				if len(rest) > 0 {
					cursor.Left(len(rest))
				}
			}
		case keys.CtrlA:
			cursor.StartOfLine()
			cursor.Right(len(PROMPT))
			pos = 0
		case keys.CtrlE:
			cursor.Right(len(cmd) - pos)
			pos = len(cmd)
		case keys.CtrlC:
			// copy to clipboard
		case keys.CtrlQ:
			FlushHistory(history)
			reset()
			return true, nil
		case keys.Enter:
			command := strings.Join(cmd, "")
			history = append(history, command)
			hpos = len(history) - 1
			cmd = []string{}
			if command == "history" {
				fmt.Println()
				for idx, cmd := range history {
					fmt.Printf("%d %s\n", idx, cmd)
				}
				ch <- "" // send empty command
			} else if strings.HasPrefix(command, "!") {
				// execute specific command
				arr := strings.Split(command, "!")
				if len(arr) > 1 {
					idxStr := strings.Trim(arr[1], " ")
					if idx, err := strconv.Atoi(idxStr); err == nil {
						if idx < len(history) {
							c := history[idx]
							ch <- c
						}
					}
				}
			} else {
				if command == "exit" || command == "quit" {
					FlushHistory(history)
					fmt.Println()
					return true, nil
				}
				ch <- command
			}
			pos = 0
			hpos += 1
		}
		return false, nil
	})
	if err != nil {
		log.Println("\nkeyboard listener failure, error:", err)
	}
}

// helper function show usage
func showUsage() {
	fmt.Println("sqlshell  commands:")
	fmt.Println("help      show this message")
	fmt.Println("history   set or show history of used commands")
	fmt.Println("!<number> execute specific command from the history")
	fmt.Println("quit      exit the sqlshell")
	fmt.Println("exit      exit the sqlshell")
	fmt.Println("set <cmd> perform set command")
	fmt.Println("          supported commands: format, connect, index, pager, limit, history")
	fmt.Println("set format=...    set output database format")
	fmt.Println("                  formats: json,pairs,rows or rows:minwidth:tabwidth:padding:padchar")
	fmt.Println("                  pairs format will show key:value pairs of single DB row (default)")
	fmt.Println("                  rows format will show record values as single DB row")
	fmt.Println("                  json format will show DB record in JSON format")
	fmt.Println("                  example: set format=rows:4:16:0")
	fmt.Println("set connect=dburi connects to provided DB uri")
	fmt.Println("                  example: set connect=sqlite:///tmp/file.db")
	fmt.Println("set history=N     limits history to N lines")
	fmt.Println("                  example: set history=1000")
	fmt.Println("set index=N       starting index from DB output")
	fmt.Println("                  example: set index=5")
	fmt.Println("set limit=N       limit cut-off from DB output")
	fmt.Println("                  example: set limit=10 (default value)")
	fmt.Println("set pager=N       shows N records per output")
	fmt.Println("                  example: set pager=2")
}

// helper function to parse DB statement
func parseDBStatement(cmd string) (string, []interface{}) {
	var args []interface{}
	cmd = strings.Replace(cmd, ";", "", -1)
	return cmd, args
}

// helper function to connect to DB
func dbConnect(dburi string) {
	db, dberr := dbInit(dburi)
	if dberr != nil {
		log.Fatal(dberr)
	}
	DB = db
}

// helper function to set DB format
func setDBFormat(format string) {
	arr := strings.Split(format, ":")
	if len(arr) > 1 {
		DBFORMAT = strings.Trim(arr[0], " ")
		if len(arr) > 1 {
			if v, e := strconv.Atoi(arr[1]); e == nil {
				MinWidth = v
			}
		}
		if len(arr) > 2 {
			if v, e := strconv.Atoi(arr[2]); e == nil {
				TabWidth = v
			}
		}
		if len(arr) > 3 {
			if v, e := strconv.Atoi(arr[3]); e == nil {
				Padding = v
			}
		}
	} else {
		DBFORMAT = strings.Trim(format, " ")
	}
}

// helper function to match DB statement
func sqlCommand(cmd string) bool {
	cmd = strings.ToLower(cmd)
	if strings.HasPrefix(cmd, "select") ||
		strings.HasPrefix(cmd, "insert") ||
		strings.HasPrefix(cmd, "update") ||
		strings.HasPrefix(cmd, "begin") ||
		strings.HasPrefix(cmd, "rollback") ||
		strings.HasPrefix(cmd, "alter") ||
		strings.HasPrefix(cmd, "commit") ||
		strings.HasPrefix(cmd, "create") ||
		strings.HasPrefix(cmd, "delete") {
		return true
	}
	return false
}

var ErrExit = errors.New("exit")

// helper function to handle commands
func cmdHandler(ch <-chan string, done <-chan bool) {
	for {
		select {
		case v := <-done:
			if v == true {
				return
			}
		case input := <-ch:
			fmt.Println("")
			// Handle the execution of the input.
			if len(input) != 0 {
				if err := execInput(input); err != nil {
					//                     log.Fprintln(os.Stderr, err)
					//                     log.Println("ERROR:", err)
					color.Error.Println("ERROR:", err)
				}
			}
			if COLOR {
				color.Info.Printf(PROMPT)
			} else {
				fmt.Printf(PROMPT)
			}
		default:
			time.Sleep(time.Duration(10) * time.Millisecond) // wait for response
		}

	}
}

// helper function to connect to ORACLE DB
func connect(uri string) error {
	return nil
}

// helper function to execute given input command
func execInput(command string) error {

	// check if empty command
	if len(command) == 0 {
		return nil
	}

	// check if we got SQL command
	if sqlCommand(command) {
		stm, args := parseDBStatement(command)
		return executeSQL(stm, args...)
	}

	// check help command
	if strings.HasPrefix(command, "help") {
		showUsage()
		return nil
	}

	// check set command
	if strings.HasPrefix(command, "set") {
		setCommand(command)
		return nil
	}

	// Remove the newline character.
	command = strings.TrimSuffix(command, "\n")

	// for ls command replace tilde with home area
	if strings.HasPrefix(command, "ls") {
		if strings.Contains(command, "~") {
			command = strings.Replace(command, "~", os.Getenv("HOME"), -1)
		}
	}

	// Split the input separate the command and the arguments.
	args := strings.Split(command, " ")

	// Check for built-in commands.
	switch args[0] {
	case "cd":
		// 'cd' to home with empty path not yet supported.
		if len(args) < 2 {
			return errors.New("path required")
		}
		// Change the directory and return the error.
		return os.Chdir(args[1])
	case "exit", "quit":
		reset()
		return ErrExit
	}

	// Prepare the command to execute.
	cmd := exec.Command(args[0], args[1:]...)

	// Set the correct output device.
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// Execute the command and return the error.
	return cmd.Run()
}

// helper set command function
/*
  set dbformat=bla
  set history=1000
  set index=1
  set limit=100
  set dbconnection=sqlite:///tmp/files.db
*/
func setCommand(input string) {
	arr := strings.Split(input, "set ")
	command := strings.Join(arr[1:], "")

	// format command
	if strings.HasPrefix(command, "format") {
		arr := strings.Split(command, "=")
		if len(arr) == 2 {
			setDBFormat(arr[1])
			fmt.Println("set DB format to", DBFORMAT)
		} else {
			fmt.Println("format  : json,cols,rows or rows:minwidth:tabwidth:padding:padchar")
			fmt.Println("example : dbformat=rows:4:16:0")
		}
		return
	}

	// connect command
	if strings.HasPrefix(command, "connect") {
		arr := strings.Split(command, "=")
		if len(arr) == 2 {
			dburi := strings.Trim(arr[1], " ")
			dbConnect(dburi)
		} else {
			fmt.Println("connect to provide DB uri, e.g. set connect sqlite:///path/file.db")
		}
		return
	}

	// history command
	if strings.HasPrefix(command, "history") {
		arr := strings.Split(command, "=")
		if len(arr) == 2 {
			s := strings.Trim(arr[1], " ")
			v, err := strconv.Atoi(s)
			if err == nil {
				HIST_LIMIT = v
			}
		} else {
			fmt.Println("set history=N, where N is number of records to keep")
		}
		return
	}

	// index command
	if strings.HasPrefix(command, "index") {
		arr := strings.Split(command, "=")
		if len(arr) == 2 {
			s := strings.Trim(arr[1], " ")
			v, err := strconv.Atoi(s)
			if err == nil {
				INDEX = v
			}
		} else {
			fmt.Println("set index=N, where N is first record from DB output to print")
		}
		return
	}

	// limit command
	if strings.HasPrefix(command, "limit") {
		arr := strings.Split(command, "=")
		if len(arr) == 2 {
			s := strings.Trim(arr[1], " ")
			v, err := strconv.Atoi(s)
			if err == nil {
				LIMIT = v
			}
		} else {
			fmt.Println("set limit=N, where N is last record from DB output to print")
		}
		return
	}

	// pager command
	if strings.HasPrefix(command, "pager") {
		arr := strings.Split(command, "=")
		if len(arr) == 2 {
			fmt.Println("Not implemented yet")
		} else {
			fmt.Println("set pager=N, where N is number of records to dump from DB output")
		}
		return
	}

	// color command
	if strings.HasPrefix(command, "color") {
		COLOR = true
		return
	}
}

// reset stdout/stderr and cursor terminal
func reset() {
	os.Stdout = nil
	os.Stderr = nil
	cursor.ClearLine()
	cursor.Show()
}
