package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/temporalio/cli/app"
)

const (
	docsPath  = "docs"
	cliFile   = "cli.md"
	filePerm  = 0644
	indexFile = "index.md"	
)

const FrontMatterTemplate = 
`---
id: {{.ID}}
title: temporal {{.Title}}{{if not .IsIndex}} {{.ID}}{{end}}
sidebar_label:{{if .IsIndex}} {{.Title}}{{else}} {{.ID}}{{end}}
description: {{.Description}}
tags:
	- cli
---

`

type FMStruct struct {
	ID string 
	Title string
	Description string
	IsIndex bool
}

var currentHeader string
var fileName string
var path string
var currentHeaderFile *os.File
var headerIndexFile string

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
	createdFiles := make(map[string]*os.File)

	// TODO: identify different option categories and print flags accordingly
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") {
			currentHeader = strings.TrimSpace(line[2:])
			path = filepath.Join(docsPath, currentHeader)
			makeFile(path, true, scanner, createdFiles)

		} else if strings.HasPrefix(line, "### ") {
			fileName = strings.TrimSpace(line[3:])
			path = filepath.Join(docsPath, currentHeader)
			// special condition for operator command file gen.
			if strings.Contains(currentHeader, "operator") {
				opPath := filepath.Join(path, fileName)
				err := os.MkdirAll(opPath, os.ModePerm)
				if err != nil {
					log.Printf("Error when trying to create directory %s: %v", path, err)
					continue
				}
				headerIndexFile := filepath.Join(opPath, indexFile)
				currentHeaderFile, err = os.Create(headerIndexFile)
				if err != nil {
					log.Printf("Error when trying to create file %s: %v", headerIndexFile, err)
					continue
				}
				createdFiles[headerIndexFile] = currentHeaderFile

				writeFrontMatter(strings.Trim(indexFile, ".md"), currentHeader, scanner, true, currentHeaderFile)
			} else {
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
			writeFrontMatter(fileName, currentHeader, scanner, false, currentHeaderFile)
		}
		} else if strings.HasPrefix(line, "#### ") {
			operatorFileName := strings.TrimSpace(line[4:])
			filePath := filepath.Join(path, fileName, operatorFileName+".md")
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
			writeFrontMatter(fileName, currentHeader, scanner, false, currentHeaderFile)
			
		} else if strings.HasPrefix(line, "**--") {
			// split into term and definition
			term, definition, found := strings.Cut(line, ":")
	
			// write to file
			term = strings.TrimSuffix(term, "=\"\"")

			// TODO: make files and separate directory and reference THAT

			if strings.Contains(term, ",") {
				makeAlias(currentHeaderFile, term)
			} else {
				writeLine(currentHeaderFile, term)
			}
			writeLine(currentHeaderFile, strings.TrimSpace(definition))
			log.Info(found)

		} else if strings.Contains(line, ">") {
			writeLine(currentHeaderFile, strings.Trim(line, ">"))
		} else {
			writeLine(currentHeaderFile, line)
		} 
	}
	// close file descriptor after for loop has completed
	readFile.Close()
	defer os.Remove(cliFile)
}

func makeFile(path string, isIndex bool, scanner *bufio.Scanner, createdFiles map[string]*os.File) {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Printf("Error when trying to create directory %s: %v", path, err)
	}
	if (isIndex) {
		headerIndexFile = filepath.Join(path, indexFile)
		currentHeaderFile, err = os.Create(headerIndexFile)
		if err != nil {
			log.Printf("Error when trying to create file %s: %v", headerIndexFile, err)
		}
		if err != nil {
			log.Printf("Error when trying to create file %s: %v", headerIndexFile, err)
		}
		createdFiles[headerIndexFile] = currentHeaderFile
		writeFrontMatter(strings.Trim(indexFile, ".md"), currentHeader, scanner, true, currentHeaderFile)
}
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

// write front matter
func writeFrontMatter (idName string, titleName string, scanner *bufio.Scanner, isIndex bool, currentHeaderFile *os.File) {
	for i := 0; i < 2; i++ {
		scanner.Scan()
	}
	descriptionTxt := scanner.Text()
	data := FMStruct{
		ID: idName,
		Title: titleName,
		Description: descriptionTxt,
		IsIndex: isIndex,
	}

	tmpl := template.Must(template.New("fm").Parse(FrontMatterTemplate))
	
	err := tmpl.ExecuteTemplate(currentHeaderFile, "fm", data)
	if err != nil {
		log.Println("Execute: ", err)
		return
	}
}

