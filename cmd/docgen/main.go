package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/temporalio/cli/app"
)

const (
	docsPath    = "docs"
	cliFile     = "cli.md"
	filePerm    = 0644
	indexFile   = "index.md"
	optionsPath = "cmd-options"
)

const FrontMatterTemplate = `---
id: {{.ID}}
title: temporal {{.Title}}{{if not .IsIndex}} {{.ID}}{{end}}
sidebar_label:{{if .IsIndex}} {{.Title}}{{else}} {{.ID}}{{end}}
description: {{.Description}}
tags:
	- cli
---
`

type FrontMatter struct {
	ID          string
	Title       string
	Description string
	IsIndex     bool
}

var currentHeader, fileName, optionFileName, operatorFileName, path, optionFilePath, headerIndexFile, aliasName string
var currentHeaderFile, currentOptionFile *os.File

// `BuildApp` takes a string and returns a `*App` and an error
func main() {
	deleteExistingFolder()

	doc, err := app.BuildApp().ToMarkdown()
	if err != nil {
		log.Fatalf("Error when trying to build app: %s", err)
	}

	err = os.WriteFile(cliFile, []byte(doc), filePerm)
	if err != nil {
		log.Fatalf("Error when trying to write markdown to %s file: %s", cliFile, err)
	}

	readFile, err := os.Open(cliFile)
	if err != nil {
		log.Fatalf("Error when trying to open %s file: %s", cliFile, err)
	}

	scanner := bufio.NewScanner(readFile)
	scanner.Split(bufio.ScanLines)
	createdFiles := make(map[string]*os.File)

	// Identify different option categories and print flags accordingly
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
			operatorFileName = strings.TrimSpace(line[4:])
			filePath := filepath.Join(path, fileName, operatorFileName+".md")
			makeFile(filePath, false, false, scanner, createdFiles)
		} else if strings.HasPrefix(line, "**--") {
			// split into term and definition
			term, definition, found := strings.Cut(line, ":")
			term = strings.TrimSuffix(term, "=\"\"")
			if strings.Contains(term, ",") {
				termArray := strings.Split(line, ",")
				optionFileName = termArray[0] + "**"
				aliasName = "Alias: **" + strings.TrimSpace(termArray[1])
			} else {
				optionFileName = term
			}
			log.Printf("string split successfully into term and definition (%v)", found)

			optionFileName = strings.TrimPrefix(optionFileName, "**--")
			optionFileName = strings.TrimSuffix(optionFileName, "**")

			optionFilePath = filepath.Join(docsPath, optionsPath, optionFileName+".md")

			termLink := "- [--" + optionFileName + "](/cli/cmd-options/" + optionFileName + ")"
			makeFile(optionFilePath, false, true, scanner, createdFiles)
			writeLine(currentHeaderFile, termLink)
			if aliasName != "" {
				aliasArray := strings.Split(aliasName, "=")
				writeLine(currentOptionFile, aliasArray[0])
				aliasName = ""
			}
			writeLine(currentOptionFile, strings.TrimSpace(definition))
		} else if strings.Contains(line, ">") {
			writeLine(currentHeaderFile, strings.Trim(line, ">"))
		} else {
			if (createdFiles[path] == currentOptionFile) || strings.Contains(line, "┌") || strings.Contains(line, "|") || strings.Contains(line, "*") || strings.Contains(line, "│"){
				writeLine(currentOptionFile, strings.TrimSpace(line))
			} else {
				writeLine(currentHeaderFile, strings.TrimSpace(line))
			}
		}
	}
	// close file descriptor after for loop has completed
	readFile.Close()
	defer os.Remove(cliFile)
}

func makeFile(path string, isIndex bool, isOptions bool, scanner *bufio.Scanner, createdFiles map[string]*os.File) {
	var err error
	if isOptions {
		err = os.MkdirAll(filepath.Join(docsPath, optionsPath), os.ModePerm)
		if err != nil {
			log.Printf("Error when trying to create options directory %s: %s", path, err)
		}
		currentOptionFile, err = os.Create(optionFilePath)
		if err != nil {
			log.Printf("Error when trying to create option file %s: %s", optionFilePath, err)
		}
		createdFiles[optionFileName] = currentOptionFile
		writeFrontMatter(strings.TrimSpace(optionFileName), "", scanner, false, currentOptionFile)

	} else if isIndex {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Printf("Error when trying to create a directory %s: %s", path, err)
		}
		headerIndexFile = filepath.Join(path, indexFile)
		currentHeaderFile, err = os.Create(headerIndexFile)
		if err != nil {
			log.Printf("Error when trying to create index file %s: %s", headerIndexFile, err)
		}
		createdFiles[headerIndexFile] = currentHeaderFile
		if !strings.Contains(path, "operator") {
			writeFrontMatter(strings.Trim(indexFile, ".md"), currentHeader, scanner, true, currentHeaderFile)
		} else {
			writeFrontMatter(strings.Trim(indexFile, ".md"), currentHeader+" "+fileName, scanner, true, currentHeaderFile)
		}
	} else {
		// check if we already created the file
		currentHeaderFile = createdFiles[path]
		if currentHeaderFile == nil {
			currentHeaderFile, err = os.Create(path)
			if err != nil {
				log.Printf("Error when trying to create non-index file %s: %s", path, err)
			}
			createdFiles[path] = currentHeaderFile
		}
		if strings.Contains(path, "operator") {
			writeFrontMatter(operatorFileName, currentHeader, scanner, false, currentHeaderFile)
			return
		}
		writeFrontMatter(fileName, currentHeader, scanner, false, currentHeaderFile)
	}
}

// It takes a file and a string, and writes the string to the file
func writeLine(file *os.File, line string) {
	_, err := file.WriteString(line + "\n")
	if err != nil {
		log.Printf("Error when trying to write to file: %s", err)
	}
}

// write front matter
func writeFrontMatter(idName string, titleName string, scanner *bufio.Scanner, isIndex bool, currentHeaderFile *os.File) {
	var descriptionTxt string
	if titleName != "" {
		for i := 0; i < 2; i++ {
			scanner.Scan()
		}
		descriptionTxt = strings.TrimSpace(scanner.Text())
	} else {
		descriptionTxt = "Definition for the " + idName + " command option."
	}

	data := FrontMatter{
		ID:          idName,
		Title:       titleName,
		Description: descriptionTxt,
		IsIndex:     isIndex,
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
		log.Printf("Folder doesn't exist.")
		return
	}
	os.RemoveAll(docsPath)
	log.Printf("deleted docs folder %s", folderinfo)
}
