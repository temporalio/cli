package helpprinter

import (
	"html/template"
	"io"
	"regexp"

	"github.com/urfave/cli/v2"
)

func HelpPrinter() func(w io.Writer, templ string, data interface{}, customFunc map[string]interface{}) {
	_helpPrinterOrig := cli.HelpPrinterCustom
	return func(w io.Writer, templ string, data interface{}, customFunc map[string]interface{}) {
		cfs := template.FuncMap{
			"markdown2Text": MarkdownToText,
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
	return removeLinks(input)
}

func removeLinks(input string) string {
	linkPattern := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	return linkPattern.ReplaceAllString(input, "$1")
}
