package tracer

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTermWriter_WriteLine(t *testing.T) {
	tests := map[string]struct {
		width   int
		height  int
		trim    bool
		content string
		want    string
	}{
		"sanity": {
			width:   10,
			height:  10,
			trim:    true,
			content: "foobarbaz",
			want:    "foobarbaz\n",
		},
		"tail trimmed content": {
			width:   10,
			height:  2,
			trim:    true,
			content: "foo\nbar",
			want:    "bar\n",
		},
		"trim wide content": {
			width:   3,
			height:  2,
			trim:    true,
			content: "foo\nbarbaz",
			want:    "",
		},
		"trimmed content doesn't cut through a line": {
			width:   3,
			height:  2,
			trim:    true,
			content: "foobarbaz",
			want:    "",
		},
		"no trimming": {
			width:   3,
			height:  2,
			trim:    false,
			content: "foobarbaz",
			want:    "foobarbaz\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := bytes.NewBufferString("") // Start with an empty string so we can test no content being written
			w := NewTermWriter(b).WithSize(tt.width, tt.height)

			_, _ = w.WriteLine(tt.content)
			err := w.Flush(tt.trim)

			require.NoError(t, err)
			require.Equalf(t, bytes.NewBufferString(tt.want), b, "flushed message doesn't match expected message")
		})
	}
}

func TestTermWriter_MultipleFlushes(t *testing.T) {
	tests := map[string]struct {
		width   int
		height  int
		trim    bool
		content []string
		want    string
	}{
		"sanity": {
			width:   10,
			height:  10,
			trim:    true,
			content: []string{"foobarbaz"},
			want:    "foobarbaz\n",
		},
		"write two single lines": {
			width:   10,
			height:  10,
			trim:    true,
			content: []string{"foo", "bar"},
			want:    fmt.Sprintf("foo\n%s%s%sbar\n", MoveCursorUp(1), AnsiMoveCursorStartLine, AnsiEraseToEnd),
		},
		"write two double lines": {
			width:   10,
			height:  10,
			trim:    true,
			content: []string{"foo\nfoo", "bar\nbar"},
			want:    fmt.Sprintf("foo\nfoo\n%s%s%sbar\nbar\n", MoveCursorUp(2), AnsiMoveCursorStartLine, AnsiEraseToEnd),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := bytes.NewBufferString("") // Start with an empty string so we can test no content being written
			w := NewTermWriter(b).WithSize(tt.width, tt.height)

			for _, s := range tt.content {
				_, _ = w.WriteLine(s)
				err := w.Flush(tt.trim)
				require.NoError(t, err)
			}

			require.Equalf(t, bytes.NewBufferString(tt.want), b, "flushed message doesn't match expected message")
		})
	}
}
