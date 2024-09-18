// Package commandsmd is built to read the markdown format described in
// temporalcli/commands.md and generate code from it.
package commandsmd

import (
	"bytes"
	_ "embed"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed commands.yml
var CommandsMarkdown []byte

type (
	// Option represents the structure of an option within option sets.
	Option struct {
		Name        string   `yaml:"name"`
		Type        string   `yaml:"type"`
		Description string   `yaml:"description"`
		Short       string   `yaml:"short,omitempty"`
		Default     string   `yaml:"default,omitempty"`
		Env         string   `yaml:"env,omitempty"`
		Required    bool     `yaml:"required,omitempty"`
		Aliases     []string `yaml:"aliases,omitempty"`
		EnumValues  []string `yaml:"enum-values,omitempty"`
	}

	// Command represents the structure of each command in the commands map.
	Command struct {
		FullName               string `yaml:"name"`
		NamePath               []string
		Summary                string `yaml:"summary"`
		Description            string `yaml:"description"`
		DescriptionPlain       string
		DescriptionHighlighted string
		HasInit                bool     `yaml:"has-init"`
		ExactArgs              int      `yaml:"exact-args"`
		MaximumArgs            int      `yaml:"maximum-args"`
		IgnoreMissingEnv       bool     `yaml:"ignores-missing-env"`
		Options                []Option `yaml:"options"`
		OptionSets             []string `yaml:"option-sets"`
	}

	// OptionSets represents the structure of option sets.
	OptionSets struct {
		Name    string   `yaml:"name"`
		Options []Option `yaml:"options"`
	}

	// Commands represents the top-level structure holding commands and option sets.
	Commands struct {
		CommandList []Command    `yaml:"commands"`
		OptionSets  []OptionSets `yaml:"option-sets"`
	}
)

func ParseMarkdownCommands() (Commands, error) {
	// Fix CRLF
	md := bytes.ReplaceAll(CommandsMarkdown, []byte("\r\n"), []byte("\n"))

	var m Commands
	err := yaml.Unmarshal(md, &m)
	if err != nil {
		return Commands{}, fmt.Errorf("failed unmarshalling yaml: %w", err)
	}

	for i, optionSet := range m.OptionSets {
		if err := m.OptionSets[i].parseSection(); err != nil {
			return Commands{}, fmt.Errorf("failed parsing option set section %q: %w", optionSet.Name, err)
		}
	}

	for i, command := range m.CommandList {
		if err := m.CommandList[i].parseSection(); err != nil {
			return Commands{}, fmt.Errorf("failed parsing command section %q: %w", command.FullName, err)
		}
	}
	return m, nil
}

var markdownLinkPattern = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
var markdownBlockCodeRegex = regexp.MustCompile("```([\\s\\S]+?)```")
var markdownInlineCodeRegex = regexp.MustCompile("`([^`]+)`")

const ansiReset = "\033[0m"
const ansiBold = "\033[1m"

func (o OptionSets) parseSection() error {
	if o.Name == "" {
		return fmt.Errorf("missing option set name")
	}

	for _, option := range o.Options {
		if err := option.parseSection(); err != nil {
			return fmt.Errorf("failed parsing option '%v': %w", option.Name, err)
		}
	}

	return nil
}

func (c *Command) parseSection() error {
	if c.FullName == "" {
		return fmt.Errorf("missing command name")
	}
	c.NamePath = strings.Split(c.FullName, " ")

	if c.Summary == "" {
		return fmt.Errorf("missing summary for command")
	}
	if c.Summary[len(c.Summary)-1] == '.' {
		return fmt.Errorf("summary should not end in a '.'")
	}

	if c.MaximumArgs != 0 && c.ExactArgs != 0 {
		return fmt.Errorf("cannot have both maximum-args and exact-args")
	}

	if c.Description == "" {
		return fmt.Errorf("missing description for command: %s", c.FullName)
	}

	// Strip links for long plain/highlighted
	c.DescriptionPlain = markdownLinkPattern.ReplaceAllString(c.Description, "$1")
	c.DescriptionHighlighted = c.DescriptionPlain

	// Highlight code for long highlighted
	c.DescriptionHighlighted = markdownBlockCodeRegex.ReplaceAllStringFunc(c.DescriptionHighlighted, func(s string) string {
		s = strings.Trim(s, "`")
		s = strings.Trim(s, " ")
		s = strings.Trim(s, "\n")
		return ansiBold + s + ansiReset
	})
	c.DescriptionHighlighted = markdownInlineCodeRegex.ReplaceAllStringFunc(c.DescriptionHighlighted, func(s string) string {
		s = strings.Trim(s, "`")
		return ansiBold + s + ansiReset
	})

	// Each option
	for _, option := range c.Options {
		if err := option.parseSection(); err != nil {
			return fmt.Errorf("failed parsing option '%v': %w", option.Name, err)
		}
	}

	return nil
}

func (o *Option) parseSection() error {
	if o.Name == "" {
		return fmt.Errorf("missing option name")
	}

	if o.Type == "" {
		return fmt.Errorf("missing option type")
	}

	if o.Description == "" {
		return fmt.Errorf("missing description for option: %s", o.Name)
	}

	if o.Env != strings.ToUpper(o.Env) {
		return fmt.Errorf("env variables must be in all caps")
	}

	if len(o.EnumValues) != 0 {
		if o.Type != "string-enum" && o.Type != "string-enum[]" {
			return fmt.Errorf("enum-values can only specified for string-enum and string-enum[] types")
		}
		// Check default enum values
		if o.Default != "" && !slices.Contains(o.EnumValues, o.Default) {
			return fmt.Errorf("default value '%s' must be one of the enum-values options %s", o.Default, o.EnumValues)
		}
	}
	return nil
}
