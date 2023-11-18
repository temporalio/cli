package trace

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestStripBytesAnsi(t *testing.T) {
	tests := map[string]struct {
		b    []byte
		want []byte
	}{
		"no ansi": {
			b:    []byte("foo"),
			want: []byte("foo"),
		},
		"moving cursor": {
			b:    []byte(fmt.Sprintf("%sfoo", MoveCursorUp(2))),
			want: []byte("foo"),
		},
		"start of line": {
			b:    []byte(fmt.Sprintf("%sfoo", AnsiMoveCursorStartLine)),
			want: []byte("foo"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equalf(t, tt.want, StripBytesAnsi(tt.b), "StripBytesAnsi(%v)", tt.b)
		})
	}
}

func TestGetBufferPrintWidth(t *testing.T) {
	tests := map[string]struct {
		b    []byte
		want int
	}{
		"sanity": {
			b:    []byte("test"),
			want: 4,
		},
		"utf8 char": {
			b:    []byte("a界c"),
			want: 3,
		},
		"ansi codes": {
			b:    []byte(color.RedString("red")),
			want: 3,
		},
		"actual output": {
			b:    []byte("\x1b[2m ╪\x1b[0m \x1b[34m▷\x1b[0m \x1b[2mcnab.CNAB_\x1b[0mRunTask \x1b[2m(5m47s ago)\x1b[0m"),
			want: 34,
		},
		"many pipes": {
			b:    []byte(" │   │   │   │   │   │   ┼ ✓ cnab.BundleExecutor_PullBundleActionWorker (119ms)"),
			want: 79,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equalf(t, tt.want, CountBytesPrintWidth(tt.b), "CountBytesPrintWidth(%q)", tt.b)
		})
	}
}

func TestNewTailBoxBoundBuffer(t *testing.T) {
	type args struct {
		buf      *bytes.Buffer
		maxLines int
		maxWidth int
	}
	tests := map[string]struct {
		args      args
		wantBuf   *bytes.Buffer
		wantLines int
	}{
		"sanity": {
			args: args{
				buf:      bytes.NewBufferString("foo\nbar\nbaz"),
				maxLines: 100,
				maxWidth: 100,
			},
			wantBuf:   bytes.NewBufferString("foo\nbar\nbaz"),
			wantLines: 3,
		},
		"zero max lines": {
			args: args{
				buf:      bytes.NewBufferString("foo\nbar\nbaz"),
				maxLines: 0,
				maxWidth: 100,
			},
			wantBuf:   bytes.NewBufferString("foo\nbar\nbaz"),
			wantLines: 3,
		},
		"too many lines": {
			args: args{
				buf:      bytes.NewBufferString("foo\nbar\nbaz"),
				maxLines: 2,
				maxWidth: 100,
			},
			wantBuf:   bytes.NewBufferString("bar\nbaz"),
			wantLines: 2,
		},
		"too wide": {
			args: args{
				buf:      bytes.NewBufferString("foo\nbar\nbaz"),
				maxLines: 100,
				maxWidth: 2,
			},
			wantBuf:   bytes.NewBufferString("foo\nbar\nbaz"),
			wantLines: 6,
		},
		"dont print unfinished lines": {
			// Since foo takes two lines and would be cut in half, we won't return it (since it might mangle some ansi codes).
			args: args{
				buf:      bytes.NewBufferString("foo\nbar\nbaz"),
				maxLines: 5,
				maxWidth: 2,
			},
			wantBuf:   bytes.NewBufferString("bar\nbaz"),
			wantLines: 4,
		},
		"actual output": {
			// TODO: consider using snapshots for these tests
			args: args{
				// 5 lines shorter than 70, 2 lines longer than 70:
				// │   ┼ ✓ cnab.bundleexecutor_experimentsisenabled (9ms)
				// │   ┼ ✓ cnab.bundleexecutor_experimentsisenabled (9ms)
				// │   ┼ ✓ cnab.bundleexecutor_experimentsisenabled (8ms)
				// │   ╪ ▷ cnab.cnabinternal_scheduleworkflowcleanup (15s ago)
				// │   │   wfid: becc1b71-1fed-4831-9ca6-05f4b212f602_51, runid: a248024d-1728-47b7-a877-ce6721aca5f7
				// │   ╪ ▷ cnab.cnab_checkworkflowauthorization (15s ago)
				// │   │   wfid: becc1b71-1fed-4831-9ca6-05f4b212f602_56, runid: 570863ad-4f66-4a03-9756-bb007eb7d3fa
				buf:      bytes.NewBufferString("\x1b[2m │   ┼\x1b[0m \x1b[32m✓\x1b[0m \x1b[2mcnab.bundleexecutor_\x1b[0mexperimentsisenabled \x1b[2m(9ms)\x1b[0m\n\x1b[2m │   ┼\x1b[0m \x1b[32m✓\x1b[0m \x1b[2mcnab.bundleexecutor_\x1b[0mexperimentsisenabled \x1b[2m(9ms)\x1b[0m\n\x1b[2m │   ┼\x1b[0m \x1b[32m✓\x1b[0m \x1b[2mcnab.bundleexecutor_\x1b[0mexperimentsisenabled \x1b[2m(8ms)\x1b[0m\n\x1b[2m │   ╪\x1b[0m \x1b[34m▷\x1b[0m \x1b[2mcnab.cnabinternal_\x1b[0mscheduleworkflowcleanup \x1b[2m(15s ago)\x1b[0m\n\x1b[2m │   │\x1b[0m   \x1b[2mwfid:\x1b[0m \x1b[34mbecc1b71-1fed-4831-9ca6-05f4b212f602_51\x1b[0m\x1b[2m, \x1b[0m\x1b[2mrunid:\x1b[0m \x1b[34ma248024d-1728-47b7-a877-ce6721aca5f7\x1b[0m\n\x1b[2m │   ╪\x1b[0m \x1b[34m▷\x1b[0m \x1b[2mcnab.cnab_\x1b[0mcheckworkflowauthorization \x1b[2m(15s ago)\x1b[0m\n\x1b[2m │   │\x1b[0m   \x1b[2mwfid:\x1b[0m \x1b[34mbecc1b71-1fed-4831-9ca6-05f4b212f602_56\x1b[0m\x1b[2m, \x1b[0m\x1b[2mrunid:\x1b[0m \x1b[34m570863ad-4f66-4a03-9756-bb007eb7d3fa\x1b[0m"),
				maxLines: 6,
				maxWidth: 70, // This should only trigger newlines on the long lines but none of the others
			},
			// Expected:
			// │   ╪ ▷ cnab.cnabinternal_scheduleworkflowcleanup (15s ago)
			// │   │   wfid: becc1b71-1fed-4831-9ca6-05f4b212f602_51, runid: a248024d-1728-47b7-a877-ce6721aca5f7
			// │   ╪ ▷ cnab.cnab_checkworkflowauthorization (15s ago)
			// │   │   wfid: becc1b71-1fed-4831-9ca6-05f4b212f602_56, runid: 570863ad-4f66-4a03-9756-bb007eb7d3fa
			wantBuf:   bytes.NewBufferString("\x1b[2m │   ╪\x1b[0m \x1b[34m▷\x1b[0m \x1b[2mcnab.cnabinternal_\x1b[0mscheduleworkflowcleanup \x1b[2m(15s ago)\x1b[0m\n\x1b[2m │   │\x1b[0m   \x1b[2mwfid:\x1b[0m \x1b[34mbecc1b71-1fed-4831-9ca6-05f4b212f602_51\x1b[0m\x1b[2m, \x1b[0m\x1b[2mrunid:\x1b[0m \x1b[34ma248024d-1728-47b7-a877-ce6721aca5f7\x1b[0m\n\x1b[2m │   ╪\x1b[0m \x1b[34m▷\x1b[0m \x1b[2mcnab.cnab_\x1b[0mcheckworkflowauthorization \x1b[2m(15s ago)\x1b[0m\n\x1b[2m │   │\x1b[0m   \x1b[2mwfid:\x1b[0m \x1b[34mbecc1b71-1fed-4831-9ca6-05f4b212f602_56\x1b[0m\x1b[2m, \x1b[0m\x1b[2mrunid:\x1b[0m \x1b[34m570863ad-4f66-4a03-9756-bb007eb7d3fa\x1b[0m"),
			wantLines: 6,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, got1 := NewTailBoxBoundBuffer(tt.args.buf, tt.args.maxLines, tt.args.maxWidth)
			assert.Equalf(t, tt.wantBuf, got, "NewTailBoxBoundBuffer(%q, %v, %v)", tt.args.buf.String(), tt.args.maxLines, tt.args.maxWidth)
			// This one helps identify mismatches
			assert.Equalf(t, tt.wantBuf.String(), got.String(), "NewTailBoxBoundBuffer(%q, %v, %v)", tt.args.buf.String(), tt.args.maxLines, tt.args.maxWidth)
			assert.Equalf(t, tt.wantLines, got1, "NewTailBoxBoundBuffer(%q, %v, %v)", tt.args.buf.String(), tt.args.maxLines, tt.args.maxWidth)
		})
	}
}
