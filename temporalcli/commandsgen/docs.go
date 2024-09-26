package commandsgen

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// index doesn't need to be generated

type DocsFile struct {
	FileName string
}

// containsOption checks if an option is in the slice
func containsOption(options []Option, option Option) bool {
	for _, opt := range options {
		// TODO: Print an error if there are duplicate names with different descriptions?
		if opt.Name == option.Name {
			return true
		}
	}
	return false
}

func GenerateDocsFiles(commands Commands) (map[string][]byte, error) {
	optionSetMap := make(map[string]OptionSets)
	for i, optionSet := range commands.OptionSets {
		optionSetMap[optionSet.Name] = commands.OptionSets[i]
	}

	w := &docWriter{fileMap: make(map[string]*bytes.Buffer), optionSetMap: optionSetMap}

	// sort by parent command (activity, batch, etc)
	for _, cmd := range commands.CommandList {
		if err := cmd.writeDoc(w); err != nil {
			return nil, fmt.Errorf("failed writing docs for command %s: %w", cmd.FullName, err)
		}
	}

	// write cmd-options
	w.writeCmdOptions()

	// Write package and imports to final buf
	// var finalBuf bytes.Buffer
	// finalBuf.WriteString("// Code generated. DO NOT EDIT.\n\n")

	// Format and return
	var finalMap = make(map[string][]byte)
	for key, buf := range w.fileMap {
		finalMap[key] = buf.Bytes()
	}
	return finalMap, nil
}

type docWriter struct {
	fileMap      map[string]*bytes.Buffer
	optionSetMap map[string]OptionSets
	optionsStack [][]Option
	allOptions   []Option
}

// func (c *docWriter) writeLinef(s string, args ...any) {
// 	// Ignore errors
// 	_, _ = c.buf.WriteString(fmt.Sprintf(s, args...) + "\n")
// }

func (c *Command) writeDoc(w *docWriter) error {
	commandLength := len(strings.Split(c.FullName, " "))
	// process options
	w.processOptions(c)
	// If this is a root command, write a new file
	if commandLength == 2 {
		w.writeCommand(c)
	} else if commandLength > 2 {
		w.writeSubcommand(c)
	}
	return nil
}

func (w *docWriter) writeCommand(c *Command) {
	fileName := strings.Split(c.FullName, " ")[1]
	w.fileMap[fileName] = &bytes.Buffer{}
	w.fileMap[fileName].WriteString("---\n")
	w.fileMap[fileName].WriteString("id: " + fileName + "\n")
	w.fileMap[fileName].WriteString("title: " + c.FullName + "\n")
	w.fileMap[fileName].WriteString("sidebar_label: " + c.FullName + "\n")
	w.fileMap[fileName].WriteString("description: " + c.Description + "\n")
	w.fileMap[fileName].WriteString("toc_max_heading_level: 4\n")
	w.fileMap[fileName].WriteString("keywords:\n") // TODO
	w.fileMap[fileName].WriteString("tags:\n")     // TODO
	w.fileMap[fileName].WriteString("---\n\n")
}

func (w *docWriter) writeSubcommand(c *Command) {
	// TODO: write options from command, parent command, and global command
	fileName := strings.Split(c.FullName, " ")[1]
	subCommand := strings.Join(strings.Split(c.FullName, " ")[2:], "")
	w.fileMap[fileName].WriteString("## " + subCommand + "\n\n")
	w.fileMap[fileName].WriteString(c.Description + "\n\n")
	w.fileMap[fileName].WriteString("Use the following options to change the behavior of this command.\n\n")
	var allOptions = make([]Option, 0)
	for _, options := range w.optionsStack {
		allOptions = append(allOptions, options...)
	}

	sort.Slice(allOptions, func(i, j int) bool {
		return allOptions[i].Name < allOptions[j].Name
	})

	for _, option := range allOptions {
		w.fileMap[fileName].WriteString(fmt.Sprintf("- [--%s](cli/cmd-options#%s)\n\n", option.Name, option.Name))
		if !containsOption(w.allOptions, option) {
			w.allOptions = append(w.allOptions, option)
		}
	}
}

func (w *docWriter) processOptions(c *Command) {
	if len(w.optionsStack) >= len(strings.Split(c.FullName, " ")) {
		w.optionsStack = w.optionsStack[:len(w.optionsStack)-1]
	}
	var options []Option
	options = append(options, c.Options...)

	// Add option sets
	for _, set := range c.OptionSets {
		// map into optionSet map
		optionSetOptions := w.optionSetMap[set].Options
		options = append(options, optionSetOptions...)
	}

	w.optionsStack = append(w.optionsStack, options)
}

// TODO: Do we even want this separate page anymore?
// Or should we just generate this description whenever it's mentioned?
// Our generator can just place this in-line on each page separately
func (w *docWriter) writeCmdOptions() {
	w.fileMap["cmd-options"] = &bytes.Buffer{}
	w.fileMap["cmd-options"].WriteString("---\n")
	w.fileMap["cmd-options"].WriteString("id: cmd-options\n")
	w.fileMap["cmd-options"].WriteString("title: Temporal CLI command options reference\n")
	w.fileMap["cmd-options"].WriteString("sidebar_label: cmd options\n")
	w.fileMap["cmd-options"].WriteString("description: Discover how to manage Temporal Workflows, from Activity Execution to Workflow Ids, using clusters, cron schedules, dynamic configurations, and logging. Perfect for developers.\n")
	w.fileMap["cmd-options"].WriteString("toc_max_heading_level: 4\n")
	w.fileMap["cmd-options"].WriteString("keywords:\n") // TODO
	w.fileMap["cmd-options"].WriteString("tags:\n")     // TODO
	w.fileMap["cmd-options"].WriteString("---\n\n")
	for _, option := range w.allOptions {
		w.fileMap["cmd-options"].WriteString(fmt.Sprintf("## %s\n\n", option.Name))
		w.fileMap["cmd-options"].WriteString(option.Description + "\n\n")
	}
}
