package trace

import (
	"bytes"
	"regexp"
)

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

// StripBytesAnsi removes all ansi codes from a byte array.
func StripBytesAnsi(b []byte) []byte {
	return re.ReplaceAllLiteral(b, []byte{})
}

// CountBytesPrintWidth counts the number of printed characters a byte array will take.
func CountBytesPrintWidth(b []byte) int {
	return len(bytes.Runes(StripBytesAnsi(b)))
}

// LineHeight returns the number of lines a string is going to take
func LineHeight(line []byte, maxWidth int) int {
	return 1 + (CountBytesPrintWidth(line)-1)/maxWidth
}

func ReverseLinesBuffer(buf *bytes.Buffer) [][]byte {
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	revBuf := make([][]byte, len(lines))

	j := 0
	// TODO: we can probably stop sooner since we're definitely not going to do more than maxLines lines
	for i := len(lines) - 1; i >= 0; i-- {
		revBuf[j] = lines[i]
		j++
	}
	return revBuf
}

// NewTailBoxBoundBuffer returns trims a buffer to fit into the box defined by maxLines and maxWidth and the number of lines printing the buffer will take.
// For no limit on lines, use maxLines = 0.
// NOTE: This is a best guess. It'll take ansi codes into account but some other chars might throw this off.
func NewTailBoxBoundBuffer(buf *bytes.Buffer, maxLines int, maxWidth int) (*bytes.Buffer, int) {
	res := make([][]byte, 0)
	lines := 0

	for _, line := range ReverseLinesBuffer(buf) {
		lineHeight := LineHeight(line, maxWidth)
		if lineHeight+lines > maxLines && maxLines > 0 {
			break
		} else {
			res = append([][]byte{line}, res...)
			lines += lineHeight
		}
	}

	return bytes.NewBuffer(bytes.Join(res, []byte{'\n'})), lines
}
