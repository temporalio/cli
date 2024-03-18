// Package commandsmd is built to read the markdown format described in
// temporalcli/commands.md and generate code from it.
package commandsmd

import (
	"bytes"
	_ "embed"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

//go:embed commands.md
var CommandsMarkdown []byte

type Command struct {
	FullName        string
	NamePath        []string
	UseSuffix       string
	Short           string
	LongPlain       string
	LongHighlighted string
	LongMarkdown    string
	OptionsSets     []CommandOptions
	HasInit         bool
	ExactArgs       int
	MaximumArgs     int
}

type CommandOptions struct {
	SetName            string
	Options            []CommandOption
	IncludeOptionsSets []string
}

type CommandOption struct {
	Name         string
	Alias        string
	DataType     string
	Desc         string
	Required     bool
	DefaultValue string
	EnumValues   []string
	EnvVar       string
	Hidden       bool
}

func ParseMarkdownCommands() ([]*Command, error) {
	// Fix CRLF
	md := bytes.ReplaceAll(CommandsMarkdown, []byte("\r\n"), []byte("\n"))

	// Split on every "### " section, ignoring the first which is markdown header
	sections := strings.Split(string(md), "\n### ")[1:]
	commands := make([]*Command, len(sections))
	for i, section := range sections {
		commands[i] = &Command{}
		if err := commands[i].parseSection(section); err != nil {
			return nil, fmt.Errorf("failed parsing section %q: %w", section[:strings.Index(section, "\n")], err)
		} else if i > 0 && commands[i-1].FullName > commands[i].FullName {
			return nil, fmt.Errorf("command %q shouldn't come before %q", commands[i-1].FullName, commands[i].FullName)
		}
	}
	return commands, nil
}

var markdownLinkPattern = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
var markdownBlockCodeRegex = regexp.MustCompile("```([\\s\\S]+?)```")
var markdownInlineCodeRegex = regexp.MustCompile("`([^`]+)`")

const ansiReset = "\033[0m"
const ansiBold = "\033[1m"

func (c *Command) parseSection(section string) error {
	// Heading
	headingEnd := strings.Index(section, "\n")
	if headingEnd == -1 {
		return fmt.Errorf("missing end of heading")
	}
	headingPieces := strings.SplitN(strings.TrimSpace(section[:headingEnd]), ":", 2)
	if len(headingPieces) != 2 {
		return fmt.Errorf("heading needs command name and short description")
	}
	c.FullName = strings.TrimSpace(headingPieces[0])
	// If there's a bracket in the name, that needs to be removed and made the use
	// suffix
	if bracketIndex := strings.Index(c.FullName, "["); bracketIndex > 0 {
		c.UseSuffix = " " + c.FullName[bracketIndex:]
		c.FullName = strings.TrimSpace(c.FullName[:bracketIndex])
	}
	c.NamePath = strings.Split(c.FullName, " ")
	c.Short = strings.TrimSpace(headingPieces[1])

	// Split into initial long description and then each options set
	subSections := strings.Split(strings.TrimSpace(section[headingEnd+1:]), "#### ")

	// Get long description, but take comment bulleted attributes off end if there
	c.LongMarkdown = strings.TrimSpace(subSections[0])
	if strings.HasSuffix(c.LongMarkdown, "-->") {
		commentStart := strings.LastIndex(c.LongMarkdown, "<!--")
		if commentStart == -1 {
			return fmt.Errorf("missing XML comment start")
		}
		bullets := strings.Split(strings.TrimSpace(strings.TrimSuffix(c.LongMarkdown[commentStart+4:], "-->")), "\n")
		c.LongMarkdown = strings.TrimSpace(c.LongMarkdown[:commentStart])
		for _, bullet := range bullets {
			bullet = strings.TrimSpace(bullet)
			switch {
			case bullet == "* has-init":
				c.HasInit = true
			case strings.HasPrefix(bullet, "* exact-args="):
				var err error
				if c.ExactArgs, err = strconv.Atoi(strings.TrimPrefix(bullet, "* exact-args=")); err != nil {
					return fmt.Errorf("invalid exact-args: %w", err)
				}
			case strings.HasPrefix(bullet, "* maximum-args="):
				var err error
				if c.MaximumArgs, err = strconv.Atoi(strings.TrimPrefix(bullet, "* maximum-args=")); err != nil {
					return fmt.Errorf("invalid maximum-args: %w", err)
				}
			default:
				return fmt.Errorf("unrecognized attribute bullet: %q", bullet)
			}
		}
		if c.MaximumArgs != 0 && c.ExactArgs != 0 {
			return fmt.Errorf("cannot have both maximum-args and exact-args")
		}
	}

	// Strip links for long plain/highlighted
	c.LongPlain = markdownLinkPattern.ReplaceAllString(c.LongMarkdown, "$1")
	c.LongHighlighted = c.LongPlain
	// Highlight code for long highlighted
	c.LongHighlighted = markdownBlockCodeRegex.ReplaceAllStringFunc(c.LongHighlighted, func(s string) string {
		s = strings.Trim(s, "`")
		s = strings.Trim(s, " ")
		s = strings.Trim(s, "\n")
		return ansiBold + s + ansiReset
	})
	c.LongHighlighted = markdownInlineCodeRegex.ReplaceAllStringFunc(c.LongHighlighted, func(s string) string {
		s = strings.Trim(s, "`")
		return ansiBold + s + ansiReset
	})

	// Each option set
	c.OptionsSets = make([]CommandOptions, len(subSections)-1)
	for i, subSection := range subSections[1:] {
		if err := c.OptionsSets[i].parseSection(strings.TrimSpace(subSection)); err != nil {
			return fmt.Errorf("failed parsing options section #%v: %w", i+1, err)
		}
	}
	return nil
}

func (c *CommandOptions) parseSection(section string) error {
	// Heading
	headingEnd := strings.Index(section, "\n")
	if headingEnd == -1 {
		return fmt.Errorf("missing end of heading")
	}
	heading := strings.TrimSpace(section[:headingEnd])
	if strings.HasPrefix(heading, "Options set for ") {
		c.SetName = strings.TrimPrefix(heading, "Options set for ")
		c.SetName = strings.TrimSuffix(c.SetName, ":")
	} else if heading != "Options" {
		return fmt.Errorf("invalid options heading")
	}
	section = strings.TrimSpace(section[headingEnd+1:])

	// Option lines
	lines := strings.Split(section, "\n")
	endOfBullets := false
	for lineIndex := 0; lineIndex < len(lines); lineIndex++ {
		line := strings.TrimSpace(lines[lineIndex])
		// Handle bullet
		if strings.HasPrefix(line, "* `") {
			if endOfBullets {
				return fmt.Errorf("got new bullet after end of bullets")
			}
			// Append each successive indented line
			for lineIndex+1 < len(lines) && strings.HasPrefix(lines[lineIndex+1], "  ") {
				line += " " + strings.TrimSpace(lines[lineIndex+1])
				lineIndex++
			}
			c.Options = append(c.Options, CommandOption{})
			if err := c.Options[len(c.Options)-1].parseBulletLine(line); err != nil {
				return fmt.Errorf("failed parsing options line end at %v: %w", lineIndex+1, err)
			}
			continue
		}
		// If we found any bullets but this isn't one, this means end of bullets
		endOfBullets = len(c.Options) > 0
		// Ignore empty
		if line == "" {
			continue
		}
		// Include
		if strings.HasPrefix(line, "Includes options set for [") {
			bracketEnd := strings.Index(line, "]")
			if bracketEnd == -1 {
				return fmt.Errorf("invalid include, missing end bracket")
			}
			c.IncludeOptionsSets = append(c.IncludeOptionsSets,
				strings.TrimPrefix(line[:bracketEnd], "Includes options set for ["))
			continue
		}
		return fmt.Errorf("unrecognized options line #%v", lineIndex+1)
	}
	return nil
}

func (c *CommandOption) parseBulletLine(bullet string) error {
	// Take off bullet
	bullet = strings.TrimPrefix(bullet, "* ")

	// Name
	if !strings.HasPrefix(bullet, "`") {
		return fmt.Errorf("missing opening backtick")
	}
	bullet = bullet[1:]
	tickEnd := strings.Index(bullet, "`")
	if tickEnd == -1 {
		return fmt.Errorf("missing ending backtick")
	} else if !strings.HasPrefix(bullet, "--") {
		return fmt.Errorf("option name %q does not have leading '--'", bullet[:tickEnd])
	}
	c.Name = strings.TrimPrefix(bullet[:tickEnd], "--")
	bullet = strings.TrimSpace(bullet[tickEnd+1:])

	// Alias
	if strings.HasPrefix(bullet, ", `") {
		bullet = strings.TrimPrefix(bullet, ", `")
		tickEnd = strings.Index(bullet, "`")
		if tickEnd == -1 {
			return fmt.Errorf("missing ending backtick")
		} else if !strings.HasPrefix(bullet, "-") {
			return fmt.Errorf("option alias %q does not have leading '-'", bullet[:tickEnd])
		}
		c.Alias = strings.TrimPrefix(bullet[:tickEnd], "-")
		bullet = strings.TrimSpace(bullet[tickEnd+1:])
	}

	// Data type
	if !strings.HasPrefix(bullet, "(") {
		return fmt.Errorf("missing data type parens")
	}
	dataTypeEnd := strings.Index(bullet, ") - ")
	if dataTypeEnd == -1 {
		return fmt.Errorf("missing data type end")
	}
	c.DataType = bullet[1:dataTypeEnd]
	bullet = strings.TrimSpace(bullet[dataTypeEnd+4:])

	// Description
	c.Desc = bullet

	// Go over trailing sentences in description to see if they're attributes and
	// take them off if they are.
	for {
		if !strings.HasSuffix(c.Desc, ".") {
			return fmt.Errorf("description doesn't end with period")
		}
		dot := strings.LastIndex(c.Desc[:len(c.Desc)-1], ". ")
		if dot == -1 {
			return nil
		}
		lastSentence := strings.TrimSpace(c.Desc[dot+1 : len(c.Desc)-1])
		switch {
		case lastSentence == "Required":
			c.Required = true
		case strings.HasPrefix(lastSentence, "Default: "):
			c.DefaultValue = strings.TrimPrefix(lastSentence, "Default: ")
		case strings.HasPrefix(lastSentence, "Options: "):
			c.EnumValues = strings.Split(strings.TrimPrefix(lastSentence, "Options: "), ", ")
		case strings.HasPrefix(lastSentence, "Env: "):
			c.EnvVar = strings.TrimPrefix(lastSentence, "Env: ")
		case lastSentence == "Hidden":
			c.Hidden = true
		default:
			return nil
		}
		c.Desc = c.Desc[:dot+1]
	}
}
