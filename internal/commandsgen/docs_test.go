package commandsgen

import "testing"

func TestEscapeMDXDescription(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "heading ID stripped",
			input:    "### Resetting activities that heartbeat {#reset-heartbeats}",
			expected: "### Resetting activities that heartbeat",
		},
		{
			name:     "angle bracket placeholder escaped",
			input:    "Use --ca-certificate <base64-encoded-cert> to set the cert.",
			expected: `Use --ca-certificate \<base64-encoded-cert\> to set the cert.`,
		},
		{
			name:     "angle brackets inside backticks left alone",
			input:    "Use `--flag <value>` to set.",
			expected: "Use `--flag <value>` to set.",
		},
		{
			name:     "JSON in single quotes escaped",
			input:    `For example: 'YourKey={"your": "value"}'.`,
			expected: `For example: 'YourKey=\{"your": "value"\}'.`,
		},
		{
			name:     "code fence content untouched",
			input:    "text\n```\n--flag <value>\n{#id}\n```\nmore text",
			expected: "text\n```\n--flag <value>\n{#id}\n```\nmore text",
		},
		{
			name:     "plain text unchanged",
			input:    "This is a normal description with no special characters.",
			expected: "This is a normal description with no special characters.",
		},
		{
			name:     "multiple placeholders on one line",
			input:    "--limit <float> --reason <string>",
			expected: `--limit \<float\> --reason \<string\>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeMDXDescription(tt.input)
			if got != tt.expected {
				t.Errorf("escapeMDXDescription(%q)\n  got:  %q\n  want: %q", tt.input, got, tt.expected)
			}
		})
	}
}
