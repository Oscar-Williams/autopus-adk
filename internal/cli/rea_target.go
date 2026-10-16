package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// reaTargetLimit caps one REA target file read: a larger file is refused
// unread, so a committed file can never make a run read without bound.
const reaTargetLimit = 1 << 20

// reaRefusal is why the CLI left a REA target file alone.
type reaRefusal int

const (
	reaRefusedSymlink    reaRefusal = iota + 1 // the path crosses a symlink, which is never followed
	reaRefusedIrregular                        // a FIFO, device, socket, or directory
	reaRefusedTooLarge                         // larger than reaTargetLimit
	reaRefusedChange                           // the REA change fails the file's structure check
	reaRefusedUnreadable                       // any other read failure
)

// reaTargetError is a REA target file the CLI did not read or write. Changed
// marks a refused change the run needed; any other refusal only means the
// file could not be inspected.
type reaTargetError struct {
	Rel     string
	Reason  reaRefusal
	Changed bool
}

// line is the one line a run prints for the file, in text's locale.
func (e *reaTargetError) line(text reaText) string {
	format := text.TargetSkipped
	if e.Changed {
		format = text.ChangeRefused
	}
	return fmt.Sprintf(format, terminalSafe(e.Rel), text.refusal(e.Reason))
}

func (e *reaTargetError) Error() string { return e.line(reaCurrentText()) }

// readReaTarget reads p's target file under dir. A missing file is nil data.
// It never follows a symlink anywhere on the path, opens only a regular file,
// and reads at most reaTargetLimit bytes; each refusal is a *reaTargetError.
func readReaTarget(dir string, p reamcp.Platform) ([]byte, error) {
	rel := reamcp.TargetRelPath(p)
	refused := func(reason reaRefusal) error { return &reaTargetError{Rel: rel, Reason: reason} }
	path := filepath.Join(dir, filepath.FromSlash(rel))
	current := filepath.Clean(dir)
	var info fs.FileInfo
	for _, part := range strings.Split(filepath.FromSlash(rel), string(os.PathSeparator)) {
		current = filepath.Join(current, part)
		var err error
		if info, err = os.Lstat(current); errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		} else if err != nil {
			return nil, refused(reaRefusedUnreadable)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, refused(reaRefusedSymlink)
		}
	}
	if !info.Mode().IsRegular() {
		return nil, refused(reaRefusedIrregular)
	}
	if info.Size() > reaTargetLimit {
		return nil, refused(reaRefusedTooLarge)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, refused(reaRefusedUnreadable)
	}
	defer file.Close()
	// The file swapped after the Lstat is not the one inspected.
	if opened, err := file.Stat(); err != nil || !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return nil, refused(reaRefusedIrregular)
	}
	data, err := io.ReadAll(io.LimitReader(file, reaTargetLimit+1))
	if err != nil {
		return nil, refused(reaRefusedUnreadable)
	}
	if len(data) > reaTargetLimit {
		return nil, refused(reaRefusedTooLarge)
	}
	return data, nil
}

// removeReaEntry removes a managed rea entry from p's target file and leaves
// any other file as it is. A file that cannot be inspected comes back as its
// read refusal. A needed removal whose result fails the file's structure
// check, or whose path crosses a symlink, is a refused change: nothing is
// written, so a file outside the project is never changed.
func removeReaEntry(dir string, p reamcp.Platform) error {
	data, err := readReaTarget(dir, p)
	if err != nil || data == nil {
		return err
	}
	rel := reamcp.TargetRelPath(p)
	if class, err := reamcp.ClassifyFile(p, data); err != nil || class != reamcp.ClassManaged {
		return nil
	}
	out, changed, err := reamcp.ApplyFile(p, data, config.ReaMCPRemove)
	if err != nil {
		return &reaTargetError{Rel: rel, Reason: reaRefusedChange, Changed: true}
	}
	if !changed {
		return nil
	}
	path, err := adapter.SafeWorkspacePath(dir, filepath.FromSlash(rel))
	if err != nil {
		return &reaTargetError{Rel: rel, Reason: reaRefusedSymlink, Changed: true}
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("REA MCP 제거 실패 %s: %w", rel, err)
	}
	if err := adapter.WriteFileIfChanged(path, out, info.Mode().Perm()); err != nil {
		return fmt.Errorf("REA MCP 제거 실패 %s: %w", rel, err)
	}
	return nil
}
