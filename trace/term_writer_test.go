package trace

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTermWriter_OneFlush(t *testing.T) {
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
			want:    "foobarbaz",
		},
		"tail trimmed content": {
			width:   10,
			height:  1,
			trim:    true,
			content: "foo\nbar",
			want:    "bar",
		},
		"trim wide content": {
			width:   3,
			height:  2,
			trim:    true,
			content: "foo\nbarbaz",
			want:    "barbaz",
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
			want:    "foobarbaz",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := bytes.NewBufferString("") // Start with an empty string so we can test no content being written
			w := NewTermWriter().WithWriter(b).WithSize(tt.width, tt.height)

			_, err := w.WriteString(tt.content)
			err = w.Flush(tt.trim)

			require.NoError(t, err)
			require.Equalf(t, b, bytes.NewBufferString(tt.want), "flushed message doesn't match expected message")
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
			want:    "foobarbaz",
		},
		"write two single lines": {
			width:   10,
			height:  10,
			trim:    true,
			content: []string{"foo", "bar"},
			want:    fmt.Sprintf("foo%s%sbar", AnsiMoveCursorStartLine, AnsiEraseToEnd),
		},
		"write two double lines": {
			width:   10,
			height:  10,
			trim:    true,
			content: []string{"foo\nfoo", "bar\nbar"},
			want:    fmt.Sprintf("foo\nfoo%s%s%sbar\nbar", MoveCursorUp(1), AnsiMoveCursorStartLine, AnsiEraseToEnd),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := bytes.NewBufferString("") // Start with an empty string so we can test no content being written
			w := NewTermWriter().WithWriter(b).WithSize(tt.width, tt.height)

			for _, s := range tt.content {
				_, err := w.WriteString(s)
				err = w.Flush(tt.trim)
				require.NoError(t, err)
			}

			require.Equalf(t, b, bytes.NewBufferString(tt.want), "flushed message doesn't match expected message")
		})
	}
}
