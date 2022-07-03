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
	"strings"
	"time"

	"atomicgo.dev/cursor"
	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
)

// PROMPT represents shell prompt
var PROMPT = "> "

// helper function to handle keyboard input
func keysHandler(ch chan<- string) {

	var pos, hpos int
	var keyList []keys.Key
	var cmd, history []string
	// start detecting user input command
	fmt.Printf(PROMPT)
	err := keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		keyList = append(keyList, key)
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
			if len(history) > 0 && hpos > 0 {
				cursor.StartOfLine()
				cursor.ClearLine()
				cmd = strings.Split(history[hpos-1], "")
				fmt.Printf(PROMPT + strings.Join(cmd, ""))
				hpos -= 1
				pos = len(cmd)
			}
		case keys.Down:
			if len(history) > 0 && hpos < len(history) {
				cursor.StartOfLine()
				cursor.ClearLine()
				cmd = strings.Split(history[hpos-1], "")
				fmt.Printf(PROMPT + strings.Join(cmd, ""))
				hpos += 1
			}
		case keys.Space:
			cursor.Right(1)
			cmd = append(cmd, " ")
			pos += 1
		case keys.Backspace:
			if pos > 0 {
				pos -= 1
				cursor.StartOfLine()
				cursor.ClearLine()
				cmd = cmd[:len(cmd)-1]
				fmt.Printf(PROMPT + strings.Join(cmd, ""))
			}
		case keys.CtrlA:
			cursor.StartOfLine()
			cursor.Right(len(PROMPT))
			pos = 0
		case keys.CtrlE:
			cursor.Right(len(cmd) - pos)
			pos = len(cmd)
		case keys.CtrlC, keys.CtrlQ, keys.CtrlX, keys.CtrlZ:
			return true, nil
		case keys.Enter:
			command := strings.Join(cmd, "")
			history = append(history, command)
			cmd = []string{}
			if command == "history" {
				fmt.Println(strings.Join(history, "\n"))
			} else if dbStatement(command) {
				stm, args := parseDBStatement(command)
				err := execute(stm, args...)
				if err != nil {
					log.Println("db error:", err)
				}
				ch <- "" // send empty command
			} else if len(command) == 0 {
				// pass
			} else {
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

// helper function to parse DB statement
func parseDBStatement(cmd string) (string, []any) {
	var args []any
	return cmd, args
}

// helper function to match DB statement
func dbStatement(cmd string) bool {
	cmd = strings.ToLower(cmd)
	if strings.HasPrefix(cmd, "select") ||
		strings.HasPrefix(cmd, "insert") ||
		strings.HasPrefix(cmd, "update") ||
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
func execInput(input string) error {
	// Remove the newline character.
	input = strings.TrimSuffix(input, "\n")

	// Split the input separate the command and the arguments.
	args := strings.Split(input, " ")

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
