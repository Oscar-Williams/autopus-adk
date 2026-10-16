package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/huh"
)

// RunNavAccessible runs form in huh's accessible mode on in and out after
// the frame: the header (none when empty), the summary block, and the
// accessible footer, one plain line each and never cut. A line that is a back
// token (navBackToken) is Esc: on a first form it prints the notice and the
// next line answers the same question; otherwise the prompt it answers reads
// an empty line, its default, and the rest of the form reads end of input, so
// huh keeps every remaining value, and the helper returns ErrNavBack. The line
// after the token stays unread for the next form.
func RunNavAccessible(form *huh.Form, frame NavFrame, in io.Reader, out io.Writer) error {
	text := navTextFor(frame.Lang)
	if frame.Header != "" {
		fmt.Fprintln(out, frame.Header)
	}
	for _, line := range frame.Summary {
		fmt.Fprintln(out, line)
	}
	footer := text.accessibleFooter
	if frame.First {
		footer = text.accessibleFooterFirst
	}
	fmt.Fprintln(out, footer)
	reader := &navLineReader{in: in, out: out, first: frame.First, notice: text.notice}
	if err := form.WithAccessible(true).WithInput(reader).WithOutput(out).Run(); err != nil {
		return err
	}
	if reader.back {
		return ErrNavBack
	}
	return nil
}

// navLineReader hands huh one input line per read and reads its own input
// one byte at a time, so every line after the current one stays unread for
// the next form: each accessible prompt builds its own scanner, which would
// otherwise buffer the answers queued behind it.
type navLineReader struct {
	in      io.Reader
	out     io.Writer
	first   bool
	notice  string
	pending []byte
	// back reports that a back token ended the form.
	back bool
}

func (r *navLineReader) Read(p []byte) (int, error) {
	if r.back && len(r.pending) == 0 {
		return 0, io.EOF
	}
	for len(r.pending) == 0 {
		line, err := r.readLine()
		if len(line) == 0 {
			return 0, err
		}
		if !navBackToken(string(line)) {
			r.pending = line
			break
		}
		if !r.first {
			// An empty line ends the token's prompt with its default. A bare
			// end of input would leave huh the prompt's last invalid answer,
			// which a Select indexes its options with.
			r.back = true
			r.pending = []byte("\n")
			break
		}
		fmt.Fprintln(r.out, r.notice)
	}
	n := copy(p, r.pending)
	r.pending = r.pending[n:]
	return n, nil
}

// readLine reads up to and including the next newline, or to the end of the
// input with the error that ended it.
func (r *navLineReader) readLine() ([]byte, error) {
	var line []byte
	one := make([]byte, 1)
	for {
		n, err := r.in.Read(one)
		if n == 1 {
			line = append(line, one[0])
			if one[0] == '\n' {
				return line, nil
			}
		}
		if err != nil {
			return line, err
		}
	}
}

// navBackToken reports whether line is the accessible back token: exactly
// b, back, or 뒤로 after trimming and ASCII case folding (REQ-TUX-06).
func navBackToken(line string) bool {
	folded := strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + ('a' - 'A')
		}
		return r
	}, strings.TrimSpace(line))
	switch folded {
	case "b", "back", "뒤로":
		return true
	}
	return false
}
