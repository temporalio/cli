package helpprinter

import (
	"html/template"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"
)

const (
	Reset = "\033[0m"
	Bold  = "\033[1m"
)

func HelpPrinter() func(w io.Writer, templ string, data interface{}, customFunc map[string]interface{}) {
	_helpPrinterOrig := cli.HelpPrinterCustom
	return func(w io.Writer, templ string, data interface{}, customFunc map[string]interface{}) {
		cfs := template.FuncMap{
			"markdown2Text": MarkdownToText,
			"trimSpace":     strings.TrimSpace,
		}

		_helpPrinterOrig(w, templ, data, cfs)
	}
}

func WithHelpTemplate(commands []*cli.Command, template string) []*cli.Command {
	for _, cmd := range commands {
		cmd.CustomHelpTemplate = template

		WithHelpTemplate(cmd.Subcommands, template)
	}

	return commands
}

func MarkdownToText(input string) string {
	input = removeLinks(input)
	input = highlightCode(input)
	return input
}

func removeLinks(input string) string {
	linkPattern := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	return linkPattern.ReplaceAllString(input, "$1")
}

func highlightCode(text string) string {
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		return text
	}

	multilineCodeBlockRegex := regexp.MustCompile("```([\\s\\S]+?)```")
	highlightedText := multilineCodeBlockRegex.ReplaceAllStringFunc(text, func(match string) string {
		codeBlock := strings.Trim(match, "`")
		codeBlock = strings.Trim(codeBlock, " ")
		codeBlock = strings.Trim(codeBlock, "\n")
		return Bold + codeBlock + Reset
	})

	inlineCodeBlockRegex := regexp.MustCompile("`([^`]+)`")
	highlightedText = inlineCodeBlockRegex.ReplaceAllStringFunc(highlightedText, func(match string) string {
		codeBlock := strings.Trim(match, "`")
		return Bold + codeBlock + Reset
	})
	return highlightedText
}
