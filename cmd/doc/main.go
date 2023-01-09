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
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/temporalio/cli/app"
)

// currently writes to one big file.
// todo: create separate files and folders.
// todo: elaborate on each file
func main() {
	// build app and convert to Markdown
	doc, err := app.BuildApp("").ToMarkdown()
	if err != nil {
		log.Fatal(err)
	}
	path := "cli.md"
	err = os.WriteFile(path, []byte(doc), 0644)
	if err != nil {
		fmt.Println(err)
	}

	// open file for scanner
	readFile, err := os.Open(path)
    if err != nil {
        fmt.Println(err)
    }

	// create scanner
	scanner := bufio.NewScanner(readFile)
	scanner.Split(bufio.ScanLines)


	// track header for file and folder creation
	var header string

	// read line
	for scanner.Scan() {
		line := scanner.Text()
		
		// directory creation
		if strings.HasPrefix(line, "## ") {
			header = strings.TrimSpace(line[2:])
			path = "/docs/" + header
			
			fmt.Println(path)
			/*error_dir := os.Mkdir("path", 0750)

			// error check
			if error_dir != nil {
				log.Fatal(error_dir)
			}

			// create index file here
			headerFile, err := os.Create(header + ".md")

			// error check
			if err != nil {
				log.Fatal(err)
			}
			
			// index file creation
			for (!strings.HasPrefix(line, "**")) {
				_, err := headerFile.WriteString(line + "\n")
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			defer headerFile.Close()


		// create files within directory
		// TODO: special case for operator commands
		} else if strings.HasPrefix(line, "###") {
			header = strings.TrimSpace(line[1:])
			headerFile, err := os.Create(header + ".md")

			// error check
			if err != nil {
				log.Fatal(err)
			}

			// file creation
			for (!strings.HasPrefix(line, "**")) {
				_, err := headerFile.WriteString(line + "\n")
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			defer headerFile.Close()

		} else {

		}*/
	}


	//close and remove big file
	readFile.Close()

	e := os.Remove("cli.md")
	if err != nil {
		log.Fatal(e)
	}
}

}