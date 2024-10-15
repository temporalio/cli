package commandsgen

import (
	"bytes"
	"fmt"
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
		usages:       commands.Usages,
	}

	// cmd-options.mdx
	w.writeCommandOptions()

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

type docWriter struct {
	fileMap      map[string]*bytes.Buffer
	optionSetMap map[string]OptionSets
	optionsStack [][]Option
	usages       Usages
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

	if len(c.Children) == 0 {
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
			w.fileMap[fileName].WriteString(fmt.Sprintf("- [--%s](/cli/cmd-options#%s)\n\n", option.Name, option.Name))
		}
	}
}

func (w *docWriter) writeCommandOptions() {
	fileName := "cmd-options"
	w.fileMap[fileName] = &bytes.Buffer{}
	w.fileMap[fileName].WriteString("---\n")
	w.fileMap[fileName].WriteString("id: " + fileName + "\n")
	w.fileMap[fileName].WriteString("title: Temporal CLI command options reference\n")
	w.fileMap[fileName].WriteString("sidebar_label: cmd options\n")
	w.fileMap[fileName].WriteString("description: Discover how to manage Temporal Workflows, from Activity Execution to Workflow Ids, using clusters, cron schedules, dynamic configurations, and logging. Perfect for developers.\n")
	w.fileMap[fileName].WriteString("toc_max_heading_level: 4\n")

	w.fileMap[fileName].WriteString("keywords:\n")
	w.fileMap[fileName].WriteString("  - " + "cli reference" + "\n")
	w.fileMap[fileName].WriteString("  - " + "command line interface cli" + "\n")
	w.fileMap[fileName].WriteString("  - " + "temporal cli" + "\n")

	w.fileMap[fileName].WriteString("tags:\n")
	w.fileMap[fileName].WriteString("  - " + "cli-reference" + "\n")
	w.fileMap[fileName].WriteString("  - " + "command-line-interface-cli" + "\n")
	w.fileMap[fileName].WriteString("  - " + "temporal-cli" + "\n")

	w.fileMap[fileName].WriteString("---\n\n")

	/////// option a
	for _, option := range w.usages.OptionUsagesByOptionDescription {
		w.fileMap[fileName].WriteString(fmt.Sprintf("## %s\n\n", option.OptionName))

		if len(option.Usages) == 1 {
			usageDescription := option.Usages[0]
			usage := usageDescription.UsageSites[0]
			w.fileMap[fileName].WriteString(usage.Option.Description + "\n\n")

			if usage.Option.Experimental {
				w.fileMap[fileName].WriteString(":::note" + "\n\n")
				w.fileMap[fileName].WriteString("Option is experimental." + "\n\n")
				w.fileMap[fileName].WriteString(":::" + "\n\n")
			}
		} else {
			for i, usageDescription := range option.Usages {
				if i > 0 {
					w.fileMap[fileName].WriteString("\n")

				}
				w.fileMap[fileName].WriteString(usageDescription.OptionDescription + "\n\n")

				for _, usage := range usageDescription.UsageSites {
					experimentalDescr := ""
					if usage.Option.Experimental {
						experimentalDescr = " (option usage is EXPERIMENTAL)"
					}
					if usage.UsageSiteType == UsageTypeCommand {
						w.fileMap[fileName].WriteString("- `" + usage.UsageSiteDescription + "`" + experimentalDescr + "\n")
					} else {
						w.fileMap[fileName].WriteString("- " + usage.UsageSiteDescription + experimentalDescr + "\n")
					}
				}

			}
		}
	}

	/////// option b

	/*

		for _, option := range w.usages.OptionUsages {
			w.fileMap[fileName].WriteString(fmt.Sprintf("## %s\n\n", option.OptionName))

			if len(option.Usages) == 1 {
				usage := option.Usages[0]
				w.fileMap[fileName].WriteString(usage.Option.Description + "\n\n")

				if usage.Option.Experimental {
					w.fileMap[fileName].WriteString(":::note" + "\n\n")
					w.fileMap[fileName].WriteString("Option is experimental and may be removed at a future date." + "\n\n")
					w.fileMap[fileName].WriteString(":::" + "\n\n")
				}
			} else {
				for _, usage := range option.Usages {
					w.fileMap[fileName].WriteString("**" + usage.UsageDescription + "**\n")
					w.fileMap[fileName].WriteString(usage.Option.Description + "\n\n")

					if usage.Option.Experimental {
						w.fileMap[fileName].WriteString(":::note" + "\n\n")
						w.fileMap[fileName].WriteString("Option is experimental and may be removed at a future date." + "\n\n")
						w.fileMap[fileName].WriteString(":::" + "\n\n")
					}
				}
			}
		}
	*/
}
