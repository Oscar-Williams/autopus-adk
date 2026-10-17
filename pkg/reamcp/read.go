package reamcp

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// TargetLimit caps one REA target file read: a larger file is refused unread,
// so a committed file can never make a run read without bound.
const TargetLimit = 1 << 20

// ReadReason is why ReadTarget left a target file unread.
type ReadReason int

const (
	ReadSymlink    ReadReason = iota + 1 // the path crosses a symlink, which is never followed
	ReadIrregular                        // a FIFO, device, socket, or directory
	ReadTooLarge                         // larger than TargetLimit
	ReadUnreadable                       // any other read failure
)

func (r ReadReason) String() string {
	switch r {
	case ReadSymlink:
		return "crosses a symlink"
	case ReadIrregular:
		return "not a regular file"
	case ReadTooLarge:
		return "larger than 1 MiB"
	}
	return "cannot be read"
}

// ReadError is a target file ReadTarget did not read.
type ReadError struct {
	Rel    string
	Reason ReadReason
}

func (e *ReadError) Error() string { return e.Rel + ": " + e.Reason.String() }

// ReadTarget reads the slash-separated project file rel under dir; a missing
// file is nil data. It never follows a symlink anywhere on the path, opens
// only a regular file, and reads at most TargetLimit bytes, so a planted
// link, FIFO, device, or oversized file can neither redirect a run outside
// the project nor stall it. Each refusal is a *ReadError. The CLI and the
// adapters share it for every REA target they inspect.
func ReadTarget(dir, rel string) ([]byte, error) {
	refused := func(reason ReadReason) error { return &ReadError{Rel: rel, Reason: reason} }
	path := filepath.Join(dir, filepath.FromSlash(rel))
	current := filepath.Clean(dir)
	var info fs.FileInfo
	for _, part := range strings.Split(filepath.FromSlash(rel), string(os.PathSeparator)) {
		current = filepath.Join(current, part)
		var err error
		if info, err = os.Lstat(current); errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		} else if err != nil {
			return nil, refused(ReadUnreadable)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, refused(ReadSymlink)
		}
	}
	if !info.Mode().IsRegular() {
		return nil, refused(ReadIrregular)
	}
	if info.Size() > TargetLimit {
		return nil, refused(ReadTooLarge)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, refused(ReadUnreadable)
	}
	defer file.Close()
	// The file swapped after the Lstat is not the one inspected.
	if opened, err := file.Stat(); err != nil || !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return nil, refused(ReadIrregular)
	}
	data, err := io.ReadAll(io.LimitReader(file, TargetLimit+1))
	if err != nil {
		return nil, refused(ReadUnreadable)
	}
	if len(data) > TargetLimit {
		return nil, refused(ReadTooLarge)
	}
	return data, nil
}
