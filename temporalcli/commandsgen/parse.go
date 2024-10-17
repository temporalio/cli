// Package commandsgen is built to read the YAML format described in
// temporalcli/commandsgen/commands.yml and generate code from it.
package commandsgen

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
var CommandsYAML []byte

type (
	// Option represents the structure of an option within option sets.
	Option struct {
		Name         string   `yaml:"name"`
		Type         string   `yaml:"type"`
		Description  string   `yaml:"description"`
		Short        string   `yaml:"short,omitempty"`
		Default      string   `yaml:"default,omitempty"`
		Env          string   `yaml:"env,omitempty"`
		Required     bool     `yaml:"required,omitempty"`
		Aliases      []string `yaml:"aliases,omitempty"`
		EnumValues   []string `yaml:"enum-values,omitempty"`
		Experimental bool     `yaml:"experimental,omitempty"`
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
		Docs                   Docs     `yaml:"docs"`
		Index                  int
		Base                   *Command
		Parent                 *Command
		Children               []*Command
		Depth                  int
		FileName               string
		SubCommandName         string
		LeafName               string
		MaxChildDepth          int
	}

	// Docs represents docs-only information that is not used in CLI generation.
	Docs struct {
		Keywords          []string `yaml:"keywords"`
		DescriptionHeader string   `yaml:"description-header"`
	}

	// OptionSets represents the structure of option sets.
	OptionSets struct {
		Name        string   `yaml:"name"`
		Description string   `yaml:"description"`
		Options     []Option `yaml:"options"`
	}

	// Commands represents the top-level structure holding commands and option sets.
	Commands struct {
		CommandList []Command    `yaml:"commands"`
		OptionSets  []OptionSets `yaml:"option-sets"`
		Usages      Usages
	}

	Usages struct {
		OptionUsages                    []OptionUsages
		OptionUsagesByOptionDescription []OptionUsagesByOptionDescription
	}

	OptionUsages struct {
		OptionName string
		UsageSites []OptionUsageSite
	}

	OptionUsageSite struct {
		Option               Option
		UsageSiteDescription string
		UsageSiteType        UsageSiteType
	}

	UsageSiteType string

	OptionUsagesByOptionDescription struct {
		OptionName string
		Usages     []OptionUsageByOptionDescription
	}

	OptionUsageByOptionDescription struct {
		OptionDescription string
		UsageSites        []OptionUsageSite
	}
)

const (
	UsageTypeCommand   UsageSiteType = "command"
	UsageTypeOptionSet UsageSiteType = "optionset"
)

func ParseCommands() (Commands, error) {
	// Fix CRLF
	md := bytes.ReplaceAll(CommandsYAML, []byte("\r\n"), []byte("\n"))

	var m Commands
	err := yaml.Unmarshal(md, &m)
	if err != nil {
		return Commands{}, fmt.Errorf("failed unmarshalling yaml: %w", err)
	}

	for i, optionSet := range m.OptionSets {
		if err := m.OptionSets[i].processSection(); err != nil {
			return Commands{}, fmt.Errorf("failed parsing option set section %q: %w", optionSet.Name, err)
		}
	}

	for i, command := range m.CommandList {
		if err := m.CommandList[i].processSection(); err != nil {
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

func (o OptionSets) processSection() error {
	if o.Name == "" {
		return fmt.Errorf("missing option set name")
	}

	for i, option := range o.Options {
		if err := o.Options[i].processSection(); err != nil {
			return fmt.Errorf("failed parsing option '%v': %w", option.Name, err)
		}
	}

	return nil
}

func (c *Command) processSection() error {
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

	if len(c.NamePath) == 2 {
		if c.Docs.Keywords == nil {
			return fmt.Errorf("missing keywords for root command: %s", c.FullName)
		}
		if c.Docs.DescriptionHeader == "" {
			return fmt.Errorf("missing description for root command: %s", c.FullName)
		}
	}

	// Strip trailing newline for description
	c.Description = strings.TrimRight(c.Description, "\n")

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
	for i, option := range c.Options {
		if err := c.Options[i].processSection(); err != nil {
			return fmt.Errorf("failed parsing option '%v': %w", option.Name, err)
		}
	}

	return nil
}

func (o *Option) processSection() error {
	if o.Name == "" {
		return fmt.Errorf("missing option name")
	}

	if o.Type == "" {
		return fmt.Errorf("missing option type")
	}

	if o.Description == "" {
		return fmt.Errorf("missing description for option: %s", o.Name)
	}
	// Strip all newline for description and trailing whitespace
	o.Description = strings.ReplaceAll(o.Description, "\n", " ")
	o.Description = strings.TrimRight(o.Description, " ")

	// Check that description ends in a "."
	if o.Description[len(o.Description)-1] != '.' {
		return fmt.Errorf("description should end in a '.'")
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
