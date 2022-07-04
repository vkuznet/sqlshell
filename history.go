package main

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
	var lines []string
	if _, err := os.Stat(path); err != nil {
		return lines
	}
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
	//     return reverse(lines)
}

// Flush writes history commands to a file
func FlushHistory(commands []string) {
	//     file, err := os.OpenFile(HistoryFile(), os.O_RDWR, 0644)
	file, err := os.Create(HistoryFile())
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
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
