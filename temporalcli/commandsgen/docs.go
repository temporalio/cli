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
		fileMap:      make(map[string]*bytes.Buffer),
		optionSetMap: optionSetMap,
	}

	// sorted ascending by full name of command (activity complete, batch list, etc)
	for _, cmd := range commands.CommandList {
		if err := cmd.writeDoc(w); err != nil {
			return nil, fmt.Errorf("failed writing docs for command %s: %w", cmd.FullName, err)
		}
	}

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
}

func (c *Command) writeDoc(w *docWriter) error {
	w.processOptions(c)

	// If this is a root command, write a new file
	if c.Depth == 1 {
		w.writeCommand(c)
	} else if c.Depth > 1 {
		w.writeSubcommand(c)
	}
	return nil
}

func (w *docWriter) writeCommand(c *Command) {
	fileName := c.FileName
	w.fileMap[fileName] = &bytes.Buffer{}
	w.fileMap[fileName].WriteString("---\n")
	w.fileMap[fileName].WriteString("id: " + fileName + "\n")
	w.fileMap[fileName].WriteString("title: " + c.FullName + "\n")
	w.fileMap[fileName].WriteString("sidebar_label: " + c.FullName + "\n")
	w.fileMap[fileName].WriteString("description: " + c.Docs.DescriptionHeader + "\n")
	w.fileMap[fileName].WriteString("toc_max_heading_level: 4\n")

	w.fileMap[fileName].WriteString("keywords:\n")
	for _, keyword := range c.Docs.Keywords {
		w.fileMap[fileName].WriteString("  - " + keyword + "\n")
	}
	// tags are the same as Keywords, but with `-` instead of ` `
	w.fileMap[fileName].WriteString("tags:\n")
	for _, keyword := range c.Docs.Keywords {
		w.fileMap[fileName].WriteString("  - " + strings.ReplaceAll(keyword, " ", "-") + "\n")
	}
	w.fileMap[fileName].WriteString("---\n\n")
}

func (w *docWriter) writeSubcommand(c *Command) {
	fileName := c.FileName
	prefix := strings.Repeat("#", c.Depth)
	w.fileMap[fileName].WriteString(prefix + " " + c.LeafName + "\n\n")
	w.fileMap[fileName].WriteString(c.Description + "\n\n")

	if c.IsLeafCommand {
		w.fileMap[fileName].WriteString("Use the following options to change the behavior of this command.\n\n")

		// gather options from command and all options aviailable from parent commands
		var allOptions = make([]Option, 0)
		for _, options := range w.optionsStack {
			allOptions = append(allOptions, options...)
		}

		// alphabetize options
		sort.Slice(allOptions, func(i, j int) bool {
			return allOptions[i].Name < allOptions[j].Name
		})

		for _, option := range allOptions {
			w.fileMap[c.FileName].WriteString(fmt.Sprintf("**--%s**\n\n", option.Name))
			w.fileMap[c.FileName].WriteString(encodeJSONExample(option.Description) + "\n\n")
			if len(option.Short) > 0 {
				w.fileMap[c.FileName].WriteString("Alias: `" + option.Short + "`\n\n")
			}
		}
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
		optionSetOptions := w.optionSetMap[set].Options
		options = append(options, optionSetOptions...)
	}

	w.optionsStack = append(w.optionsStack, options)
}

func encodeJSONExample(v string) string {
	// example: 'YourKey={"your": "value"}'
	// results in an mdx acorn rendering error
	// and wrapping in backticks lets it render
	re := regexp.MustCompile(`('[a-zA-Z0-9]*={.*}')`)
	v = re.ReplaceAllString(v, "`$1`")
	return v
}
