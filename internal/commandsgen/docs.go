package commandsgen

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

func GenerateDocsFiles(commands Commands) (map[string][]byte, error) {
	optionSetMap := make(map[string]OptionSets)
	for i, optionSet := range commands.OptionSets {
		optionSetMap[optionSet.Name] = commands.OptionSets[i]
	}

	w := &docWriter{
		fileMap:        make(map[string]*bytes.Buffer),
		optionSetMap:   optionSetMap,
		allCommands:    commands.CommandList,
		globalFlagsMap: make(map[string]map[string]Option),
	}

	// sorted ascending by full name of command (activity complete, batch list, etc)
	for _, cmd := range commands.CommandList {
		if err := cmd.writeDoc(w); err != nil {
			return nil, fmt.Errorf("failed writing docs for command %s: %w", cmd.FullName, err)
		}
	}

	// Write global flags section once at the end of each file
	w.writeGlobalFlagsSections()

	// Format and return
	var finalMap = make(map[string][]byte)
	for key, buf := range w.fileMap {
		finalMap[key] = buf.Bytes()
	}
	return finalMap, nil
}

type docWriter struct {
	allCommands    []Command
	fileMap        map[string]*bytes.Buffer
	optionSetMap   map[string]OptionSets
	optionsStack   [][]Option
	globalFlagsMap map[string]map[string]Option // fileName -> optionName -> Option
}

func (c *Command) writeDoc(w *docWriter) error {
	w.processOptions(c)

	// If this is a root command, write a new file
	depth := c.depth()
	if depth == 1 {
		w.writeCommand(c)
	} else if depth > 1 {
		w.writeSubcommand(c)
	}
	return nil
}

func (w *docWriter) writeCommand(c *Command) {
	fileName := c.fileName()
	w.fileMap[fileName] = &bytes.Buffer{}
	w.fileMap[fileName].WriteString("---\n")
	w.fileMap[fileName].WriteString("id: " + fileName + "\n")
	w.fileMap[fileName].WriteString("title: Temporal CLI " + fileName + " command reference\n")
	w.fileMap[fileName].WriteString("sidebar_label: " + fileName + "\n")
	w.fileMap[fileName].WriteString("description: " + c.Docs.DescriptionHeader + "\n")
	w.fileMap[fileName].WriteString("toc_max_heading_level: 4\n")

	w.fileMap[fileName].WriteString("keywords:\n")
	for _, keyword := range c.Docs.Keywords {
		w.fileMap[fileName].WriteString("  - " + keyword + "\n")
	}
	w.fileMap[fileName].WriteString("tags:\n")
	for _, tag := range c.Docs.Tags {
		w.fileMap[fileName].WriteString("  - " + tag + "\n")
	}
	w.fileMap[fileName].WriteString("---")
	w.fileMap[fileName].WriteString("\n\n")
	w.fileMap[fileName].WriteString("{/* NOTE: This is an auto-generated file. Any edit to this file will be overwritten.\n")
	w.fileMap[fileName].WriteString("This file is generated from https://github.com/temporalio/cli/blob/main/internal/commandsgen/commands.yml via internal/cmd/gen-docs */}\n")
}

func (w *docWriter) writeSubcommand(c *Command) {
	fileName := c.fileName()
	prefix := strings.Repeat("#", c.depth())
	w.fileMap[fileName].WriteString(prefix + " " + c.leafName() + "\n\n")
	w.fileMap[fileName].WriteString(c.Description + "\n\n")

	if w.isLeafCommand(c) {
		// gather options from command and all options available from parent commands
		var options = make([]Option, 0)
		var globalOptions = make([]Option, 0)
		for i, o := range w.optionsStack {
			if i == len(w.optionsStack)-1 {
				options = append(options, o...)
			} else {
				globalOptions = append(globalOptions, o...)
			}
		}

		// alphabetize options
		sort.Slice(options, func(i, j int) bool {
			return options[i].Name < options[j].Name
		})

		// Only write command-specific flags here
		if len(options) > 0 {
			w.fileMap[fileName].WriteString("Use the following options to change the behavior of this command.\n\n")
			w.writeOptionsTable(options, c)
		}

		// Collect global flags for later (deduplicated)
		w.collectGlobalFlags(fileName, globalOptions)
	}
}

func (w *docWriter) writeOptionsTable(options []Option, c *Command) {
	if len(options) == 0 {
		return
	}

	fileName := c.fileName()
	buf := w.fileMap[fileName]

	// Command-specific flags: 3 columns (no Default)
	buf.WriteString("| Flag | Required | Description |\n")
	buf.WriteString("|------|----------|-------------|\n")

	for _, o := range options {
		w.writeOptionRow(buf, o, false)
	}
	buf.WriteString("\n")
}

func (w *docWriter) writeOptionRow(buf *bytes.Buffer, o Option, includeDefault bool) {
	// Flag name column
	flagName := fmt.Sprintf("`--%s`", o.Name)
	if len(o.Short) > 0 {
		flagName += fmt.Sprintf(", `-%s`", o.Short)
	}

	// Required column
	required := "No"
	if o.Required {
		required = "Yes"
	}

	// Description column - starts with data type
	optionType := o.Type
	if o.DisplayType != "" {
		optionType = o.DisplayType
	}
	description := fmt.Sprintf("**%s** %s", optionType, encodeJSONExample(o.Description))
	if len(o.EnumValues) > 0 {
		description += fmt.Sprintf(" Accepted values: %s.", strings.Join(o.EnumValues, ", "))
	}
	if o.Experimental {
		description += " _(Experimental)_"
	}
	// Escape pipes in description for table compatibility
	description = strings.ReplaceAll(description, "|", "\\|")

	if includeDefault {
		// Default column
		defaultVal := ""
		if len(o.Default) > 0 {
			defaultVal = fmt.Sprintf("`%s`", o.Default)
		}
		buf.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", flagName, required, description, defaultVal))
	} else {
		buf.WriteString(fmt.Sprintf("| %s | %s | %s |\n", flagName, required, description))
	}
}

func (w *docWriter) collectGlobalFlags(fileName string, options []Option) {
	if w.globalFlagsMap[fileName] == nil {
		w.globalFlagsMap[fileName] = make(map[string]Option)
	}
	for _, o := range options {
		// Only add if not already present (deduplication)
		if _, exists := w.globalFlagsMap[fileName][o.Name]; !exists {
			w.globalFlagsMap[fileName][o.Name] = o
		}
	}
}

func (w *docWriter) writeGlobalFlagsSections() {
	for fileName, optionsMap := range w.globalFlagsMap {
		if len(optionsMap) == 0 {
			continue
		}

		// Convert map to slice and sort
		options := make([]Option, 0, len(optionsMap))
		for _, o := range optionsMap {
			options = append(options, o)
		}
		sort.Slice(options, func(i, j int) bool {
			return options[i].Name < options[j].Name
		})

		buf := w.fileMap[fileName]
		buf.WriteString("## Global Flags\n\n")
		buf.WriteString("The following options can be used with any command.\n\n")
		// Global flags: 4 columns (with Default)
		buf.WriteString("| Flag | Required | Description | Default |\n")
		buf.WriteString("|------|----------|-------------|--------|\n")

		for _, o := range options {
			w.writeOptionRow(buf, o, true)
		}
		buf.WriteString("\n")
	}
}

func (w *docWriter) processOptions(c *Command) {
	// Pop options from stack if we are moving up a level
	if len(w.optionsStack) >= len(strings.Split(c.FullName, " ")) {
		w.optionsStack = w.optionsStack[:len(w.optionsStack)-1]
	}
	var options []Option
	options = append(options, c.Options...)

	// Maintain stack of options available from parent commands
	for _, set := range c.OptionSets {
		optionSet, ok := w.optionSetMap[set]
		if !ok {
			panic(fmt.Sprintf("invalid option set %v used", set))
		}
		optionSetOptions := optionSet.Options
		options = append(options, optionSetOptions...)
	}

	w.optionsStack = append(w.optionsStack, options)
}

func (w *docWriter) isLeafCommand(c *Command) bool {
	for _, maybeSubCmd := range w.allCommands {
		if maybeSubCmd.isSubCommand(c) {
			return false
		}
	}
	return true
}

func encodeJSONExample(v string) string {
	// example: 'YourKey={"your": "value"}'
	// results in an mdx acorn rendering error
	// and wrapping in backticks lets it render
	re := regexp.MustCompile(`('[a-zA-Z0-9]*={.*}')`)
	v = re.ReplaceAllString(v, "`$1`")
	return v
}
