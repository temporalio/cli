package app_test

import (
	"github.com/temporalio/cli/helpprinter"
)

func (s *cliAppSuite) TestMarkdownToText() {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input: `The ` + "`" + `temporal workflow describe` + "`" + ` command shows information about a given [Workflow Execution](/concepts/what-is-a-workflow-execution).
	This information can be used to locate Workflow Executions that weren't able to run successfully.

	Use the command options listed below to change the information returned by this command.
	Make sure to write the command in this format:
	` + "`" + `temporal workflow describe [command options]` + "`" + ``,
			expected: `The ` + "`" + `temporal workflow describe` + "`" + ` command shows information about a given Workflow Execution.
	This information can be used to locate Workflow Executions that weren't able to run successfully.

	Use the command options listed below to change the information returned by this command.
	Make sure to write the command in this format:
	` + "`" + `temporal workflow describe [command options]` + "`" + ``,
		},
	}

	for _, test := range tests {
		s.Equal(test.expected, helpprinter.MarkdownToText(test.input))
	}
}
