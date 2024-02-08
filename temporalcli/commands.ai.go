package temporalcli

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/temporalio/cli/temporalcli/ai"
)

func InitializeAICommands() *cobra.Command {
	aiCmd := &cobra.Command{
		Use:   "ai",
		Short: "AI related commands",
		Long:  `All commands related to AI functionalities.`,
	}

	askAiCmd := &cobra.Command{
		Use:   "ask",
		Short: "Ask a question",
		Long:  `Ask any question and have it echoed back.`,
		Run: func(cmd *cobra.Command, args []string) {
			question := strings.Join(args, " ")
			ai.SendQuestion(question)
		},
	}

	aiCmd.AddCommand(askAiCmd)

	return aiCmd
}
