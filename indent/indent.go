package indent

import (
	"bytes"
	"io"
	"strings"

	"github.com/muesli/reflow/ansi"
)

type IndentFunc func(w io.Writer)

type Writer struct {
	Indent     uint
	IndentFunc IndentFunc

	ansiWriter *ansi.Writer
	buf        bytes.Buffer
	skipIndent bool
	ansi       bool
}

func NewWriter(indent uint, indentFunc IndentFunc) *Writer {
	w := &Writer{
		Indent:     indent,
		IndentFunc: indentFunc,
	}
	w.ansiWriter = &ansi.Writer{
		Forward: &w.buf,
	}
	return w
}

func NewWriterPipe(forward io.Writer, indent uint, indentFunc IndentFunc) *Writer {
	return &Writer{
		Indent:     indent,
		IndentFunc: indentFunc,
		ansiWriter: &ansi.Writer{
			Forward: forward,
		},
	}
}

// Bytes is shorthand for declaring a new default indent-writer instance,
// used to immediately indent a byte slice.
func Bytes(b []byte, indent uint) []byte {
	f := NewWriter(indent, nil)
	_, _ = f.Write(b)

	return f.Bytes()
}

// String is shorthand for declaring a new default indent-writer instance,
// used to immediately indent a string.
func String(s string, indent uint) string {
	return string(Bytes([]byte(s), indent))
}

const ESCAPE_ITERM_IMAGE_IN = "\x1b]1337;"
const ESCAPE_ITERM_IMAGE_OUT = '\x07'

func IsImageEscapeSequence(s string) bool {
	return strings.Contains(s, ESCAPE_ITERM_IMAGE_IN)
}

// Write is used to write content to the indent buffer.
func (w *Writer) Write(b []byte) (int, error) {

	var escape_seq = ""
	for _, c := range string(b) {
		if c == '\x1B' {
			escape_seq = ""
			escape_seq += string(c)
			// ANSI escape sequence
			w.ansi = true
		} else if c == ESCAPE_ITERM_IMAGE_OUT {
			w.ansi = false
			escape_seq = ""
		} else if w.ansi {
			escape_seq += string(c)
			if (c >= 0x41 && c <= 0x5a) || (c >= 0x61 && c <= 0x7a) {
				// ANSI sequence terminated
				w.ansi = false
			}
		} else if IsImageEscapeSequence(escape_seq) {
				w.ansi = false
		} else {
			if !w.skipIndent {
				w.ansiWriter.ResetAnsi()
				if w.IndentFunc != nil {
					for i := 0; i < int(w.Indent); i++ {
						w.IndentFunc(w.ansiWriter)
					}
				} else {
					_, err := w.ansiWriter.Write([]byte(strings.Repeat(" ", int(w.Indent))))
					if err != nil {
						return 0, err
					}
				}

				w.skipIndent = true
				w.ansiWriter.RestoreAnsi()
			}

			if c == '\n' {
				// end of current line
				w.skipIndent = false
			}
		}

		_, err := w.ansiWriter.Write([]byte(string(c)))
		if err != nil {
			return 0, err
		}
	}

	return len(b), nil
}

// Bytes returns the indented result as a byte slice.
func (w *Writer) Bytes() []byte {
	return w.buf.Bytes()
}

// String returns the indented result as a string.
func (w *Writer) String() string {
	return w.buf.String()
}
