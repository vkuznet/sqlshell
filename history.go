package main

// history module
// Copyright (c) 2022 - Valentin Kuznetsov <vkuznet@gmail.com>

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
)

func HistoryFile() string {
	udir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(udir, ".sqlshell_history")
	return path
}

// helper function to reverse array of strings
func reverse(cmds []string) []string {
	for i, j := 0, len(cmds)-1; i < j; i, j = i+1, j-1 {
		cmds[i], cmds[j] = cmds[j], cmds[i]
	}
	return cmds
}

// Read reads content from history file
func ReadHistory() []string {
	path := HistoryFile()
	var commands []string
	if _, err := os.Stat(path); err != nil {
		return commands
	}
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		commands = append(commands, scanner.Text())
	}
	if len(commands) > HIST_LIMIT {
		idx := len(commands) - HIST_LIMIT
		commands = commands[idx:]
	}
	return commands
}

// Flush writes history commands to a file
func FlushHistory(commands []string) {
	fname := HistoryFile()
	file, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	//     file, err := os.Create(HistoryFile())
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	if len(commands) > HIST_LIMIT {
		idx := len(commands) - HIST_LIMIT
		commands = commands[idx:]
	}
	for _, cmd := range commands {
		if cmd == "" {
			continue
		}
		_, err := file.WriteString(cmd + "\n")
		if err != nil {
			log.Fatal(err)
		}
	}
}
