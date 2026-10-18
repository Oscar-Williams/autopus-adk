package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// reaTargetLimit caps one REA target file read (reamcp.TargetLimit).
const reaTargetLimit = reamcp.TargetLimit

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

// readReaTarget reads p's target file under dir through reamcp.ReadTarget: a
// missing file is nil data, and a symlink anywhere on the path, a file that
// is not regular, or one larger than reaTargetLimit comes back as a
// *reaTargetError naming the reason.
func readReaTarget(dir string, p reamcp.Platform) ([]byte, error) {
	rel := reamcp.TargetRelPath(p)
	data, err := reamcp.ReadTarget(dir, rel)
	var refused *reamcp.ReadError
	if errors.As(err, &refused) {
		return nil, &reaTargetError{Rel: rel, Reason: reaRefusalOf(refused.Reason)}
	}
	return data, err
}

// reaRefusalOf maps a guarded-read reason onto the CLI's refusal text.
func reaRefusalOf(reason reamcp.ReadReason) reaRefusal {
	switch reason {
	case reamcp.ReadSymlink:
		return reaRefusedSymlink
	case reamcp.ReadIrregular:
		return reaRefusedIrregular
	case reamcp.ReadTooLarge:
		return reaRefusedTooLarge
	}
	return reaRefusedUnreadable
}

// removeReaEntry removes a managed rea entry from p's target file, deleting
// the Antigravity plugin file when the entry was all it held, and leaves any
// other file as it is. A file that cannot be inspected comes back as its
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
	if reamcp.RemovesFile(p, out) {
		// The Antigravity plugin file held the managed entry alone.
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("REA MCP 제거 실패 %s: %w", rel, err)
		}
		return nil
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
