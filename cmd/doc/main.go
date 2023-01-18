package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/temporalio/cli/app"
)

const (
	docsPath  = "docs"
	cliFile   = "cli.md"
	filePerm  = 0644
	indexFile = "index.md"
)

// `BuildApp` takes a string and returns a `*App` and an error
func main() {
	doc, err := app.BuildApp("").ToMarkdown()
	if err != nil {
		log.Fatalf("Error when trying to build app: %v", err)
	}

	err = os.WriteFile(cliFile, []byte(doc), filePerm)
	if err != nil {
		log.Fatalf("Error when trying to write markdown to %s file: %v", cliFile, err)
	}

	readFile, err := os.Open(cliFile)
	if err != nil {
		log.Fatalf("Error when trying to open %s file: %v", cliFile, err)
	}
	scanner := bufio.NewScanner(readFile)
	scanner.Split(bufio.ScanLines)

	var currentHeader string
	var currentHeaderFile *os.File
	createdFiles := make(map[string]*os.File)

	// TODO: identify different option categories and print flags accordingly
	// TODO: rework what is written to the string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") {
			currentHeader = strings.TrimSpace(line[2:])
			path := filepath.Join(docsPath, currentHeader)

			err := os.MkdirAll(path, os.ModePerm)
			if err != nil {
				log.Printf("Error when trying to create directory %s: %v", path, err)
				continue
			}

			headerIndexFile := filepath.Join(path, indexFile)
			currentHeaderFile, err = os.Create(headerIndexFile)
			if err != nil {
				log.Printf("Error when trying to create file %s: %v", headerIndexFile, err)
				continue
			}
			createdFiles[headerIndexFile] = currentHeaderFile
			writeLine(currentHeaderFile, line)

		} else if strings.HasPrefix(line, "### ") {
			path := filepath.Join(docsPath, currentHeader)
			fileName := strings.TrimSpace(line[3:])

			filePath := filepath.Join(path, fileName+".md")
			// check if already created file
			currentHeaderFile = createdFiles[filePath]
			if currentHeaderFile == nil {
				currentHeaderFile, err = os.Create(filePath)
				if err != nil {
					log.Printf("Error when trying to create file %s: %v", filePath, err)
					continue
				}
				createdFiles[filePath] = currentHeaderFile
			}
			writeLine(currentHeaderFile, line)
		} else if strings.HasPrefix(line, "**--") {
			// split into term and definition
			term, definition, found := strings.Cut(line, ":")
	
			// write to file
			term = strings.TrimSuffix(term, "=\"\"")

			// TODO: separate terms and aliases
			if strings.Contains(term, ",") {
				makeAlias(currentHeaderFile, term)
			} else {
				writeLine(currentHeaderFile, term)
			}
			writeLine(currentHeaderFile, strings.TrimSpace(definition))
			log.Info(found)

		} else {
			writeLine(currentHeaderFile, line)
		} 
	}
	// close file descriptor after for loop has completed
	readFile.Close()
	defer os.Remove(cliFile)
}

// It takes a file and a string, and writes the string to the file
func writeLine(file *os.File, line string) {
	_, err := file.WriteString(line + "\n")
	if err != nil {
		log.Printf("Error when trying to write to file: %v", err)
	}
}

// separates aliases from terms
func makeAlias(file *os.File, line string) {
	termArray := strings.Split(line, ",")
	writeLine(file, termArray[0] + "**")
	writeLine(file, "Alias: **" + strings.TrimSpace(termArray[1]))
}

