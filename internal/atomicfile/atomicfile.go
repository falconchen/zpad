// Package atomicfile writes a file through a temporary file in the same
// directory, then renames it into place, so a crash cannot leave a
// half-written file.
package atomicfile

import (
	"io"
	"os"
	"path/filepath"
)

// Write writes data to dir/name with mode 0600, for settings and session
// files.
func Write(dir, name string, data []byte) error {
	return WriteMode(dir, name, data, 0o600)
}

// WriteMode writes data to dir/name with mode.
func WriteMode(dir, name string, data []byte, mode os.FileMode) error {
	return WriteTo(dir, name, mode, func(w io.Writer) error {
		_, err := w.Write(data)
		return err
	})
}

// WriteTo writes what fn writes to dir/name with mode, a chunk at a time
// through a temporary file, so that a large text is written without a copy
// of it in memory.
func WriteTo(dir, name string, mode os.FileMode, fn func(io.Writer) error) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "."+name+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		if tmpName != "" {
			_ = os.Remove(tmpName)
		}
	}()

	if err := fn(tmp); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if mode == 0 {
		mode = 0o644
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		return err
	}
	if err := os.Rename(tmpName, filepath.Join(dir, name)); err != nil {
		return err
	}
	tmpName = ""
	return nil
}

// WriteFile writes data to path, keeping the file's directory and using
// mode, or the mode of the file already there.
func WriteFile(path string, data []byte, mode os.FileMode) error {
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	return WriteMode(filepath.Dir(path), filepath.Base(path), data, mode)
}
