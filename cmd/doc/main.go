// The MIT License
//
// Copyright (c) 2022 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/temporalio/cli/app"
)

// todo: create separate files and folders.
// todo: elaborate on each file
func main() {
	// build app and convert to Markdown
	doc, err := app.BuildApp("").ToMarkdown()
	fatal_check(err)

	path := "cli.md"
	err = os.WriteFile(path, []byte(doc), 0644)
	print_check(err)

	// open file for scanner
	readFile, err := os.Open(path)
    print_check(err)

	// create scanner
	scanner := bufio.NewScanner(readFile)
	scanner.Split(bufio.ScanLines)


	// track header for file and folder creation
	var header string
	var headerFile *os.File

	// read line
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") {
			header = strings.TrimSpace(line[2:])
			path_docs := "docs/" + header

			//create directory
			err := os.MkdirAll(path_docs, os.ModePerm)
			print_check(err)
			
			// create index file here
			headerFile, err = os.Create(filepath.Join(path_docs, "index.md"))
			print_check(err)

		} else if strings.HasPrefix(line, "### "){
			path_docs := "docs/" + header
			fileName := strings.TrimSpace(line[3:])

			//create file within directory
			headerFile, err = os.Create(filepath.Join(path_docs, fileName + ".md"))
			print_check(err)

		} else if !strings.HasPrefix(line, "# ") {
			_, err := headerFile.WriteString(line + "\n")
			print_check(err)
		} else {
			continue
		}

	}
	//close and remove big file
	readFile.Close()
	e := os.Remove("cli.md")
	fatal_check(e)
}

// I got sick of putting these code blocks everywhere, so now they're functions.
func fatal_check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func print_check(e error) {
	if e != nil {
		log.Println(e)
	}
}