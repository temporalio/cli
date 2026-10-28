package commandsgen

import (
	"strings"
	"testing"
)

func TestEscapeMDXDescription(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "bare angle-bracket placeholder is escaped",
			in:   "Pass the cert via --ca-certificate <base64-encoded-cert> to authenticate.",
			want: `Pass the cert via --ca-certificate \<base64-encoded-cert\> to authenticate.`,
		},
		{
			name: "placeholder with separators and colon is escaped",
			in:   "Set --limit <requests_per_second:float> and --reason <reason_string>.",
			want: `Set --limit \<requests_per_second:float\> and --reason \<reason_string\>.`,
		},
		{
			name: "angle brackets inside a fenced code block are left untouched",
			in:   "Example:\n\n```\ntemporal x --cert <base64-encoded-cert>\n```\n",
			want: "Example:\n\n```\ntemporal x --cert <base64-encoded-cert>\n```\n",
		},
		{
			name: "angle brackets inside an inline code span are left untouched",
			in:   "Use `--cert <base64-encoded-cert>` here but escape <key> there.",
			want: "Use `--cert <base64-encoded-cert>` here but escape \\<key\\> there.",
		},
		{
			name: "custom heading id is converted to MDX comment form",
			in:   "### Resetting activities that heartbeat {#reset-heartbeats}\n\nSee [details](#reset-heartbeats).",
			want: "### Resetting activities that heartbeat {/* #reset-heartbeats */}\n\nSee [details](#reset-heartbeats).",
		},
		{
			name: "non-heading line with {#id}-like text is not treated as a heading",
			in:   "See [details](#reset-heartbeats).",
			want: "See [details](#reset-heartbeats).",
		},
		{
			name: "bare single-quoted JSON has its braces escaped",
			in:   `Provide default '{"some-key": "some-value"}' inline.`,
			want: `Provide default '\{"some-key": "some-value"\}' inline.`,
		},
		{
			name: "single-quoted JSON inside a fence is left untouched",
			in:   "```\ntemporal x --input '{\"some-key\": \"some-value\"}'\n```",
			want: "```\ntemporal x --input '{\"some-key\": \"some-value\"}'\n```",
		},
		{
			name: "plain prose is unchanged",
			in:   "This command does something useful with no special characters.",
			want: "This command does something useful with no special characters.",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := escapeMDXDescription(tc.in); got != tc.want {
				t.Errorf("escapeMDXDescription()\n got: %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestEncodeJSONExample(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "key=json example is wrapped",
			in:   `Example: 'YourKey={"your": "value"}'.`,
			want: "Example: `'YourKey={\"your\": \"value\"}'`.",
		},
		{
			name: "standalone json example is wrapped",
			in:   `Example: '{"some-key": "some-value"}'.`,
			want: "Example: `'{\"some-key\": \"some-value\"}'`.",
		},
		{
			name: "nested json example is wrapped",
			in:   `Example: '{"a": {"b": "c"}}'.`,
			want: "Example: `'{\"a\": {\"b\": \"c\"}}'`.",
		},
		{
			name: "no json is unchanged",
			in:   "A plain description without JSON.",
			want: "A plain description without JSON.",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := encodeJSONExample(tc.in); got != tc.want {
				t.Errorf("encodeJSONExample()\n got: %q\nwant: %q", got, tc.want)
			}
		})
	}
}

// splitFixture is a minimal command tree with a "cloud" root, mirroring how the
// cloud CLI extension is structured, used to exercise -subdir splitting.
const splitFixture = `
commands:
  - name: cloud
    summary: Cloud
    description: Manage Temporal Cloud.
  - name: cloud thing
    summary: Thing
    description: |
      Manage things.
    docs:
      keywords:
        - thing
      description-header: Manage things
      tags:
        - Cloud
  - name: cloud thing sub
    summary: Sub
    description: |
      Do a thing with a bare <placeholder> outside code.
    options:
      - name: flag
        type: string
        description: A flag.
`

func generateSplitFixture(t *testing.T, subdirs []string) map[string][]byte {
	t.Helper()
	cmds, err := ParseCommands([]byte(splitFixture))
	if err != nil {
		t.Fatalf("ParseCommands: %v", err)
	}
	docs, err := GenerateDocsFiles(cmds, subdirs)
	if err != nil {
		t.Fatalf("GenerateDocsFiles: %v", err)
	}
	return docs
}

func TestGenerateDocsFilesSubdir(t *testing.T) {
	docs := generateSplitFixture(t, []string{"cloud"})

	if _, ok := docs["cloud/thing"]; !ok {
		t.Errorf("expected split file cloud/thing, got keys: %v", keys(docs))
	}
	if _, ok := docs["thing"]; ok {
		t.Errorf("did not expect a flat 'thing' file when splitting cloud")
	}
	if _, ok := docs["cloud"]; ok {
		t.Errorf("split parent 'cloud' should not produce a standalone file")
	}
	// Index pages are hand-maintained on the docs site, so gen-docs never emits them.
	if _, ok := docs["cloud/index"]; ok {
		t.Errorf("gen-docs should not emit a cloud/index page")
	}
	if _, ok := docs["index"]; ok {
		t.Errorf("gen-docs should not emit a top-level index page")
	}

	// The deeper subcommand's description is appended to its parent file and
	// must be MDX-escaped there.
	thing := string(docs["cloud/thing"])
	if !strings.Contains(thing, `\<placeholder\>`) {
		t.Errorf("expected escaped placeholder in cloud/thing, got:\n%s", thing)
	}
	if !strings.Contains(thing, "## sub") {
		t.Errorf("expected 'sub' heading in cloud/thing, got:\n%s", thing)
	}
}

func keys(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// parentOptionsFixture has a parent command that defines its own options and
// can be run directly ("thing run", like "workflow reset"), alongside a parent
// that only groups subcommands ("thing group").
const parentOptionsFixture = `
commands:
  - name: app
    summary: App
    description: App.
  - name: app thing
    summary: Thing
    description: Manage things.
    docs:
      keywords:
        - thing
      description-header: Manage things
      tags:
        - Things
  - name: app thing group
    summary: Group
    description: Group related subcommands.
  - name: app thing group leaf
    summary: Leaf
    description: A leaf under a grouping parent.
    options:
      - name: leaf-flag
        type: string
        description: A leaf flag.
  - name: app thing run
    summary: Run
    description: Run a thing.
    options:
      - name: target
        type: string
        description: The thing to run.
  - name: app thing run extra
    summary: Extra
    description: Run a thing, then do something extra.
    options:
      - name: extra-flag
        type: string
        description: An extra flag.
`

func TestGenerateDocsFilesParentOptions(t *testing.T) {
	cmds, err := ParseCommands([]byte(parentOptionsFixture))
	if err != nil {
		t.Fatalf("ParseCommands: %v", err)
	}

	cases := []struct {
		name      string
		subdirs   []string
		file      string
		run       string
		extra     string
		group     string
		groupLeaf string
	}{
		{name: "single file", file: "thing", run: "## run", extra: "### extra", group: "## group", groupLeaf: "### leaf"},
		{name: "split subdirectory", subdirs: []string{"app"}, file: "app/thing", run: "## run", extra: "### run extra", group: "## group", groupLeaf: "### group leaf"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			docs, err := GenerateDocsFiles(cmds, tc.subdirs)
			if err != nil {
				t.Fatalf("GenerateDocsFiles: %v", err)
			}
			doc, ok := docs[tc.file]
			if !ok {
				t.Fatalf("expected file %q, got keys: %v", tc.file, keys(docs))
			}

			// A parent with its own options gets an option table under its heading.
			run := section(t, string(doc), tc.run, tc.extra)
			if !strings.Contains(run, "`--target`") {
				t.Errorf("expected --target in the run section, got:\n%s", run)
			}
			// Its options are persistent flags, so the section says they reach subcommands.
			if !strings.Contains(run, "this command and all of its subcommands") {
				t.Errorf("expected the run section to say its options apply to subcommands, got:\n%s", run)
			}
			extra := section(t, string(doc), tc.extra, "")
			if !strings.Contains(extra, "`--extra-flag`") {
				t.Errorf("expected --extra-flag in the extra section, got:\n%s", extra)
			}

			// A parent that only groups subcommands still gets no table.
			group := section(t, string(doc), tc.group, tc.groupLeaf)
			if strings.Contains(group, "| Flag |") {
				t.Errorf("expected no option table in the group section, got:\n%s", group)
			}
		})
	}
}

// section returns the text from the start heading up to the end heading, or to
// the end of the document when end is empty.
func section(t *testing.T, doc, start, end string) string {
	t.Helper()
	i := strings.Index(doc, start+"\n")
	if i < 0 {
		t.Fatalf("heading %q not found in:\n%s", start, doc)
	}
	rest := doc[i:]
	if end == "" {
		return rest
	}
	j := strings.Index(rest, end+"\n")
	if j < 0 {
		t.Fatalf("heading %q not found after %q in:\n%s", end, start, doc)
	}
	return rest[:j]
}
