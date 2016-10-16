package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestParsingFileNames(t *testing.T) {
	f, err := os.Open("support/files.txt")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer f.Close()

	files := []string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fname := filepath.Base(line)
		parts := strings.Split(fname, ".")
		for i, part := range parts {
			time.Now().Year()
			matched, err := regexp.MatchString(reString, part)
			if err != nil {
				t.Fatalf("failed to match file part: %v", err)
			}
			if matched {
				parts[i] = ""
			}
		}
		files = append(files, strings.Join(parts, " "))
	}
	fmt.Println(files)
	// Maybe want to nuke 4 digit numbers from before today's year? But what about 1984???
	if err := scanner.Err(); err != nil {
		t.Fatalf("%v", err)
	}
	t.Fatal()
}
