# Probe A1/A2: huh v1.0.0 Esc, Ctrl+C, accessible back token

- Date: 2026-10-16. Base: `feat/tuiux-001` at 8ec13935 (huh v1.0.0, bubbletea v1.3.6, go1.27.0 darwin/arm64).
- How: a throwaway test `internal/cli/tui/zz_tuiux_probe_test.go` (package tui), run with `go test ./internal/cli/tui -run TestProbeA -v -count=1`, then deleted. The source is kept below so the run can be repeated.
- P1 sends `tea.KeyMsg` straight to `Form.Update`. P2 runs a real `tea.Program` over an `io.Pipe` with a wrapper model (Esc -> back flag and `tea.Quit`, first step -> notice and keep running, everything else -> the form). P3 runs the accessible form over a line reader that turns a `b`/`back`/`뒤로` line into EOF for the rest of the form.

## Output

```text
P1 default form after esc: state=0 (0=normal,1=completed,2=aborted) value="b"
P1 default form after ctrl+c: state=2
P1 default Quit keys=[ctrl+c]
P2 esc                    value="b" err=probe: back back=true note=false state=0
P2 ctrl+c                 value="b" err=user aborted back=false note=false state=2
P2 down+enter             value="b" err=probe: back back=true note=false state=0
P2 enter                  value="b" err=<nil> back=false note=false state=1
P2 esc-first-then-enter   value="b" err=<nil> back=false note=true state=1
P3 input="b\n" value="b" back=true err=<nil> out="Q \n1. A\n2. B\n3. C\nEnter a number between 1 and 3: \n\n"
P3 input="2\n" value="b" back=false err=<nil> out="Q \n1. A\n2. B\n3. C\nEnter a number between 1 and 3: \n"
P3 input="\n" value="b" back=false err=<nil> out="Q \n1. A\n2. B\n3. C\nEnter a number between 1 and 3: \n"
P3 input="뒤로\n3\n" value="b" back=true err=<nil> out="Q \n1. A\n2. B\n3. C\nEnter a number between 1 and 3: \n\n"
P3 group a="ko" b="en" back=true err=<nil>
P2b one-chunk down then enter: value="c" back=false state=1 err=<nil>
P3b accessible 3: value="c"
ok  	github.com/insajin/autopus-adk/internal/cli/tui	1.790s
```

## Reading

- P1: with the default KeyMap, Esc leaves the form in StateNormal (it is not an abort) and Ctrl+C aborts. `Quit` is `[ctrl+c]` only, so the wizard comment "ctrl+c or esc" (`wizard_steps.go:65`) does not hold for v1.0.0.
- P2: a wrapper model tells Esc (back flag) from Ctrl+C (form StateAborted -> `huh.ErrUserAborted`) on the same form. On the first step Esc only sets the notice and the form keeps running; the following Enter completes it. `GetFocusedField`, `GetFiltering` and `WithShowHelp` compile against v1.0.0.
- P2 `down+enter` sent one byte per write: the lone `` arrived alone and was read as Esc. P2b sent `[B` in one write and it was read as Down. A split escape sequence (slow link) can therefore look like Esc. Back is recoverable (values are restored and preselected), so this is a risk, not a blocker.
- P3: accessible mode has no back of its own (`Form.runAccessible` drops field errors, `PromptInt` re-prompts on non-numbers). The line reader gives one: the token line ends the form with every value unchanged, and the line after it (`3
`) is left for the next form. In a two-select group the first answer landed before the token (`a="ko"`), so the helper must restore the state it saved before the form.

## Probe source

```go
package tui

// Throwaway probe for SPEC-TUIUX-001 A1/A2. Deleted after the run.

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

var errProbeBack = errors.New("probe: back")

type probeNav struct {
	form  *huh.Form
	first bool
	back  bool
	note  bool
}

func (m *probeNav) Init() tea.Cmd { return m.form.Init() }

func (m *probeNav) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.Type == tea.KeyEsc {
		if f, ok := m.form.GetFocusedField().(interface{ GetFiltering() bool }); ok && f.GetFiltering() {
			goto pass
		}
		if m.first {
			m.note = true
			return m, nil
		}
		m.back = true
		return m, tea.Quit
	}
pass:
	model, cmd := m.form.Update(msg)
	m.form = model.(*huh.Form)
	if m.form.State != huh.StateNormal {
		return m, tea.Quit
	}
	return m, cmd
}

func (m *probeNav) View() string {
	if m.back || m.form.State != huh.StateNormal {
		return ""
	}
	return "HEADER\n" + m.form.View() + "\nFOOTER"
}

func probeSelect(v *string) *huh.Form {
	return huh.NewForm(huh.NewGroup(huh.NewSelect[string]().Title("Q").
		Options(huh.NewOption("A", "a"), huh.NewOption("B", "b"), huh.NewOption("C", "c")).Value(v))).
		WithShowHelp(false)
}

func runProbeTUI(t *testing.T, input string, first bool) (string, error, *probeNav) {
	v := "b"
	m := &probeNav{form: probeSelect(&v), first: first}
	pr, pw := io.Pipe()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	go func() {
		time.Sleep(200 * time.Millisecond)
		for _, b := range []byte(input) {
			_, _ = pw.Write([]byte{b})
			time.Sleep(150 * time.Millisecond)
		}
	}()
	var out bytes.Buffer
	_, err := tea.NewProgram(m, tea.WithInput(pr), tea.WithOutput(&out), tea.WithContext(ctx), tea.WithoutSignalHandler()).Run()
	_ = pw.Close()
	if err == nil && m.form.State == huh.StateAborted {
		err = huh.ErrUserAborted
	}
	if m.back {
		err = errProbeBack
	}
	return v, err, m
}

func TestProbeA1_DefaultEscIsNotAbort(t *testing.T) {
	v := "b"
	f := probeSelect(&v)
	f.Init()
	f.Update(tea.KeyMsg{Type: tea.KeyEsc})
	t.Logf("P1 default form after esc: state=%d (0=normal,1=completed,2=aborted) value=%q", f.State, v)
	g := probeSelect(&v)
	g.Init()
	g.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	t.Logf("P1 default form after ctrl+c: state=%d", g.State)
	t.Logf("P1 default Quit keys=%v", huh.NewDefaultKeyMap().Quit.Keys())
}

func TestProbeA1_WrapperDistinguishesEscFromCtrlC(t *testing.T) {
	for _, tc := range []struct {
		name, input string
		first       bool
	}{{"esc", "\x1b", false}, {"ctrl+c", "\x03", false}, {"down+enter", "\x1b[B\r", false}, {"enter", "\r", false}, {"esc-first-then-enter", "\x1b\r", true}} {
		v, err, m := runProbeTUI(t, tc.input, tc.first)
		t.Logf("P2 %-22s value=%q err=%v back=%v note=%v state=%d", tc.name, v, err, m.back, m.note, m.form.State)
	}
}

// probeBackReader passes whole lines through and turns a back-token line
// into EOF for the rest of the form.
type probeBackReader struct {
	in      io.Reader
	pending []byte
	back    bool
}

func (r *probeBackReader) Read(p []byte) (int, error) {
	if r.back {
		return 0, io.EOF
	}
	if len(r.pending) == 0 {
		var line []byte
		one := make([]byte, 1)
		for {
			n, err := r.in.Read(one)
			if n == 1 {
				line = append(line, one[0])
				if one[0] == '\n' {
					break
				}
			}
			if err != nil {
				if len(line) == 0 {
					return 0, err
				}
				break
			}
		}
		if s := strings.ToLower(strings.TrimSpace(string(line))); s == "b" || s == "back" || s == "뒤로" {
			r.back = true
			return 0, io.EOF
		}
		r.pending = line
	}
	n := copy(p, r.pending)
	r.pending = r.pending[n:]
	return n, nil
}

func TestProbeA2_AccessibleBackToken(t *testing.T) {
	for _, input := range []string{"b\n", "2\n", "\n", "뒤로\n3\n"} {
		v := "b"
		src := bufio.NewReader(strings.NewReader(input))
		r := &probeBackReader{in: src}
		var out bytes.Buffer
		err := probeSelect(&v).WithAccessible(true).WithInput(r).WithOutput(&out).Run()
		t.Logf("P3 input=%q value=%q back=%v err=%v out=%q", input, v, r.back, err, out.String())
	}
	// Two selects in one group (the language step shape): the first answer
	// lands before the token, so the helper must restore its snapshot.
	a, b := "en", "en"
	r := &probeBackReader{in: strings.NewReader("2\nb\n")}
	var out bytes.Buffer
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().Title("first").Options(huh.NewOption("en", "en"), huh.NewOption("ko", "ko")).Value(&a),
		huh.NewSelect[string]().Title("second").Options(huh.NewOption("en", "en"), huh.NewOption("ko", "ko")).Value(&b)))
	err := form.WithAccessible(true).WithInput(r).WithOutput(&out).Run()
	t.Logf("P3 group a=%q b=%q back=%v err=%v", a, b, r.back, err)
}

func TestProbeA1_ChunkedArrowIsNotEsc(t *testing.T) {
	v := "b"
	m := &probeNav{form: probeSelect(&v)}
	pr, pw := io.Pipe()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	go func() {
		time.Sleep(200 * time.Millisecond)
		_, _ = pw.Write([]byte("\x1b[B"))
		time.Sleep(150 * time.Millisecond)
		_, _ = pw.Write([]byte("\r"))
	}()
	_, err := tea.NewProgram(m, tea.WithInput(pr), tea.WithOutput(io.Discard), tea.WithContext(ctx), tea.WithoutSignalHandler()).Run()
	_ = pw.Close()
	t.Logf("P2b one-chunk down then enter: value=%q back=%v state=%d err=%v", v, m.back, m.form.State, err)
	w := "b"
	var out bytes.Buffer
	_ = probeSelect(&w).WithAccessible(true).WithInput(&probeBackReader{in: strings.NewReader("3\n")}).WithOutput(&out).Run()
	t.Logf("P3b accessible 3: value=%q", w)
}
```
