package commandsgen

import "testing"

func TestGenerateDeprecationBox(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:    "default message when empty",
			message: "",
			expected: "```\n" +
				"+-----------------------------------------------------------------------------+\n" +
				"| CAUTION: This command is deprecated and will be removed in a later release. |\n" +
				"+-----------------------------------------------------------------------------+\n" +
				"```\n\n",
		},
		{
			name:    "custom message",
			message: "Use the new API instead.",
			expected: "```\n" +
				"+-----------------------------------+\n" +
				"| CAUTION: Use the new API instead. |\n" +
				"+-----------------------------------+\n" +
				"```\n\n",
		},
		{
			name:    "short custom message",
			message: "Removed.",
			expected: "```\n" +
				"+-------------------+\n" +
				"| CAUTION: Removed. |\n" +
				"+-------------------+\n" +
				"```\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateDeprecationBox(tt.message)
			if got != tt.expected {
				t.Errorf("generateDeprecationBox(%q) =\n%q\nwant:\n%q", tt.message, got, tt.expected)
			}
		})
	}
}
