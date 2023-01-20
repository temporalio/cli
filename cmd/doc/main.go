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
	optionsPath = "cmd-options"	
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

var currentHeader, fileName, optionFileName, path, optionFilePath, headerIndexFile  string
var currentHeaderFile, currentOptionFile *os.File

// `BuildApp` takes a string and returns a `*App` and an error
func main() {
	deleteExistingFolder()

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
			makeFile(path, true, false, scanner, createdFiles)
		} else if strings.HasPrefix(line, "### ") {
			fileName = strings.TrimSpace(line[3:])
			path = filepath.Join(docsPath, currentHeader)
			if strings.Contains(currentHeader, "operator") {
				opPath := filepath.Join(path, fileName)
				makeFile(opPath, true, false, scanner, createdFiles)
			} else {
				filePath := filepath.Join(path, fileName+".md")
				makeFile(filePath, false, false, scanner, createdFiles)
			}
		} else if strings.HasPrefix(line, "#### ") {
			operatorFileName := strings.TrimSpace(line[4:])
			filePath := filepath.Join(path, fileName, operatorFileName+".md")
			makeFile(filePath, false, false, scanner, createdFiles)
			
		} else if strings.HasPrefix(line, "**--") {
			// split into term and definition
			term, definition, found := strings.Cut(line, ":")
			term = strings.TrimSuffix(term, "=\"\"")
			if strings.Contains(term, ",") {
				makeAlias(currentHeaderFile, term)
			} else {
				writeLine(currentHeaderFile, term)
				optionFileName = term
			}
			writeLine(currentHeaderFile, strings.TrimSpace(definition))
			log.Info("string split successfully into term and definition (%v)",found)

			optionFileName = strings.TrimPrefix(optionFileName, "**--")
			optionFileName = strings.TrimSuffix(optionFileName, "**")

			optionFilePath = filepath.Join(docsPath, optionsPath, optionFileName+".md")

			makeFile(optionFilePath, false, true, scanner, createdFiles)

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

func makeFile(path string, isIndex bool, isOptions bool, scanner *bufio.Scanner, createdFiles map[string]*os.File) {
	var err error
	if (isOptions) {
		err = os.MkdirAll(filepath.Join(docsPath, optionsPath), os.ModePerm)
		if err != nil {
			log.Printf("Error when trying to create options directory %s: %v", path, err)
		}
		currentOptionFile, err = os.Create(optionFilePath)
		if err != nil {
			log.Printf("Error when trying to create option file %s: %v", optionFilePath, err)
		}
		createdFiles[optionFileName] = currentOptionFile
		//writeFrontMatter(optionFileName, "", scanner, false, currentOptionFile)
			
		} else if (isIndex) {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Printf("Error when trying to create a directory %s: %v", path, err)
		}
		headerIndexFile = filepath.Join(path, indexFile)
		currentHeaderFile, err = os.Create(headerIndexFile)
		if err != nil {
			log.Printf("Error when trying to create index file %s: %v", headerIndexFile, err)
		}
		createdFiles[headerIndexFile] = currentHeaderFile
		writeFrontMatter(strings.Trim(indexFile, ".md"), currentHeader, scanner, true, currentHeaderFile)
	} else {
		// check if we already created the file
		currentHeaderFile = createdFiles[path]
		if currentHeaderFile == nil {
			currentHeaderFile, err = os.Create(path)
			if err != nil {
				log.Printf("Error when trying to create non-index file %s: %v", path, err)
			}
			createdFiles[path] = currentHeaderFile
		}
		writeFrontMatter(fileName, currentHeader, scanner, false, currentHeaderFile)
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
	optionFileName = termArray[0]
}

// write front matter
func writeFrontMatter(idName string, titleName string, scanner *bufio.Scanner, isIndex bool, currentHeaderFile *os.File) {
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

func deleteExistingFolder() {
	folderinfo, err := os.Stat(docsPath)
	if os.IsNotExist(err) {
		log.Info("Folder doesn't exist.")
		return
	}
	os.RemoveAll(docsPath)
	log.Println("deleted docs folder %v", folderinfo)
}

