package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/temporalio/cli/temporalcli/commandsmd"
)

func main() {
	commands, err := commandsmd.ParseMarkdownCommands()
	if err != nil {
		fmt.Println("Error parsing commands:", err)
		return
	}

	funcMap := template.FuncMap{
		"join": strings.Join,
		"last": func(ss []string) string {
			return ss[len(ss)-1]
		},
	}

	templateStr := `---
id: {{.NamePath}}
title: temporal {{.FullName}}
sidebar_label: {{.NamePath | last}}
description: {{.Short}}
tags:
	- cli reference
	- temporal cli
	- {{.NamePath}}
---

{{.LongMarkdown}}

` + "`" + `temporal {{.FullName}}{{.UseSuffix}}` + "`" + `

Use the following command options to change the information returned by this command.

{{range .OptionsSets}}
{{range .Options}}
- [{{.Name}}](/cli/cmd-options/{{.Name}})
{{end}}
{{end}}
`

	tmpl, err := template.New("cli_reference").Funcs(funcMap).Parse(templateStr)
	if err != nil {
		fmt.Println("Error creating template:", err)
		return
	}

	docsDir := "docs"
	err = os.MkdirAll(docsDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating docs directory:", err)
		return
	}

	for _, command := range commands {
		dirPath := filepath.Join(docsDir, strings.Join(command.NamePath[:len(command.NamePath)-1], "/"))
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dirPath, err)
			continue
		}

		fileName := command.NamePath[len(command.NamePath)-1] + ".md"
		filePath := filepath.Join(dirPath, fileName)

		file, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", filePath, err)
			continue
		}
		defer file.Close()

		err = tmpl.Execute(file, command)
		if err != nil {
			fmt.Printf("Error executing template for %s: %v\n", command.FullName, err)
			continue
		}

		fmt.Printf("Generated documentation for %s -> %s\n", command.FullName, filePath)
	}
}
