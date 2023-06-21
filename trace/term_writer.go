package trace

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
)

const (
	AnsiMoveCursorUp        = "\x1b[%dA"
	AnsiMoveCursorStartLine = "\x1b[0G"
	AnsiEraseToEnd          = "\x1b[0J"
)

func MoveCursorUp(lines int) string {
	return fmt.Sprintf(AnsiMoveCursorUp, lines)
}

type TermWriter struct {
	// out is the writer to write to
	out io.Writer

	buf       *bytes.Buffer
	mtx       sync.Mutex
	lineCount int

	termWidth  int
	termHeight int
}

// NewTermWriter returns a new TermWriter set to output to Stdout.
// TermWriter is a stateful writer designed to print into a terminal window by limiting the number of lines printed what fits and clearing them on new outputs.
func NewTermWriter() *TermWriter {
	w := &TermWriter{buf: new(bytes.Buffer)}
	return w.WithWriter(io.Writer(os.Stdout))
}

// WithWriter sets the writer for TermWriter.
func (w *TermWriter) WithWriter(out io.Writer) *TermWriter {
	w.out = out
	return w
}

// WithSize sets the size of TermWriter to the desired width and height.
func (w *TermWriter) WithSize(width, height int) *TermWriter {
	if width <= 0 {
		panic(fmt.Errorf("TermWriter cannot have width %d", width))
	}
	if height <= 0 {
		panic(fmt.Errorf("TermWriter cannot have height %d", width))
	}
	w.termHeight = height
	w.termWidth = width
	return w
}

func (w *TermWriter) GetSize() (int, int) {
	return w.termWidth, w.termHeight
}

// WithTerminalSize sets the size of TermWriter to that of the terminal.
func (w *TermWriter) WithTerminalSize() *TermWriter {
	termWidth, termHeight := getTerminalSize()
	return w.WithSize(termWidth, termHeight)
}

// Write save the contents of buf to the writer b. The only errors returned are ones encountered while writing to the underlying buffer.
// TODO: Consider merging it into Flush since we might always want to write and flush. Alternatively, we can pass the writer to the Sprint functions and write (but we might run into issues if the normal writing is interrupted and the interrupt writing starts).
func (w *TermWriter) Write(buf []byte) (n int, err error) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	return w.buf.Write(buf)
}

// WriteString writes a string into TermWriter.
func (w *TermWriter) WriteString(s string) (n int, err error) {
	return w.Write([]byte(s))
}

// clearLines prints the ansi codes to clear a given number of lines.
func (w *TermWriter) clearLines() error {
	if w.lineCount == 0 {
		return nil
	}
	b := bytes.NewBuffer([]byte{})
	if w.lineCount > 1 {
		_, _ = b.Write([]byte(MoveCursorUp(w.lineCount - 1)))
	}
	_, _ = b.Write([]byte(AnsiMoveCursorStartLine))
	_, _ = b.Write([]byte(AnsiEraseToEnd))
	_, err := w.out.Write(b.Bytes())
	return err
}

// Flush writes to the out and resets the buffer. It should be called after the last call to Write to ensure that any data buffered in the TermWriter is written to output.
// Any incomplete escape sequence at the end is considered complete for formatting purposes.
// An error is returned if the contents of the buffer cannot be written to the underlying output stream.
func (w *TermWriter) Flush(trim bool) error {
	if w.termWidth <= 0 {
		return fmt.Errorf("TermWriter cannot flush without a valid width (current: %d)", w.termWidth)
	}

	w.mtx.Lock()
	defer w.mtx.Unlock()

	// Do nothing if buffer is empty.
	if len(w.buf.Bytes()) == 0 {
		return nil
	}

	maxLines := 0
	if trim {
		maxLines = w.termHeight
	}
	// Tail the buffer (if necessary) and count the number of lines
	r, lines := NewTailBoxBoundBuffer(w.buf, maxLines, w.termWidth)

	// Clear the last printed lines and store the new amount of lines that are going to be printed.
	if err := w.clearLines(); err != nil {
		return err
	}
	w.lineCount = lines

	_, err := w.out.Write(r.Bytes())
	w.buf.Reset()
	return err
}
