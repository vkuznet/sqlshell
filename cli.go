package main

// sqlshell: a better replacement for DB shell(s)
// Copyright (c) 2015 - Valentin Kuznetsov <vkuznet@gmail.com>
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
)

// PROMPT represents shell prompt
var PROMPT = "> "

// DBFORMAT defines how to print DB records, e.g. columns or rows
var DBFORMAT string

// MinWidth used by tabwriter
var MinWidth int = 4

// TabWidth used by tabwriter
var TabWidth int = 16

// Padding used by tabwriter
var Padding int = 0

// helper function to handle keyboard input
func keysHandler(ch chan<- string) {
	var pos, hpos int
	var cmd []string
	history := ReadHistory()
	if len(history) > 0 {
		hpos = len(history)
	}

	// start detecting user input command
	fmt.Printf(PROMPT)
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
			fmt.Printf(PROMPT + strings.Join(cmd, ""))
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
				fmt.Printf(PROMPT + strings.Join(cmd, ""))
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
				fmt.Printf(PROMPT + strings.Join(cmd, ""))
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
				fmt.Printf(PROMPT + strings.Join(cmd, ""))
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
		case keys.CtrlC, keys.CtrlQ, keys.CtrlX, keys.CtrlZ:
			FlushHistory(history)
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
				if command == "exit" {
					FlushHistory(history)
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
	fmt.Println("sqlshell commands:")
	fmt.Println("help     - this message")
	fmt.Println("history  - show history of used commands")
	fmt.Println("           use !<number> to execute specific command from the history")
	fmt.Println("dbformat - set output database format")
	fmt.Println("           supported formats: json,pairs,rows or rows:minwidth:tabwidth:padding:padchar")
	fmt.Println("           pairs format will show key:value pairs of single DB row (default)")
	fmt.Println("           rows format will show record values as single DB row")
	fmt.Println("           json format will show DB record in JSON format")
	fmt.Println("           example : dbformat=rows:4:16:0")
}

// helper function to parse DB statement
func parseDBStatement(cmd string) (string, []interface{}) {
	var args []interface{}
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
func dbStatement(cmd string) bool {
	cmd = strings.ToLower(cmd)
	if strings.HasPrefix(cmd, "select") ||
		strings.HasPrefix(cmd, "insert") ||
		strings.HasPrefix(cmd, "update") ||
		strings.HasPrefix(cmd, "begin") ||
		strings.HasPrefix(cmd, "rollback") ||
		strings.HasPrefix(cmd, "commit") ||
		strings.HasPrefix(cmd, "delete") {
		return true
	}
	return false
}

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
					fmt.Fprintln(os.Stderr, err)
				}
			}
			fmt.Printf(PROMPT)
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
	if dbStatement(command) {
		stm, args := parseDBStatement(command)
		err := execute(stm, args...)
		if err != nil {
			log.Println("db error:", err)
		}
		return nil
	}

	// check help command
	if strings.HasPrefix(command, "help") {
		showUsage()
		return nil
	}

	// check dbformat command
	if strings.HasPrefix(command, "dbformat") {
		arr := strings.Split(command, "=")
		if len(arr) == 2 {
			setDBFormat(arr[1])
			fmt.Println("set DB format to", DBFORMAT)
		} else {
			fmt.Println("dbformat: json,cols,rows or rows:minwidth:tabwidth:padding:padchar")
			fmt.Println("Example : dbformat=rows:4:16:0")
		}
		return nil
	}

	// check dbconnect command
	if strings.HasPrefix(command, "dbconnect") {
		arr := strings.Split(command, "=")
		if len(arr) == 2 {
			dburi := strings.Trim(arr[1], " ")
			dbConnect(dburi)
		}
		return nil
	}
	// Remove the newline character.
	command = strings.TrimSuffix(command, "\n")

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
	case "exit":
		os.Exit(0)
	}

	// Prepare the command to execute.
	cmd := exec.Command(args[0], args[1:]...)

	// Set the correct output device.
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// Execute the command and return the error.
	return cmd.Run()
}
