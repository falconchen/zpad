// Package draft keeps recovery copies of the text of unsaved documents, so
// that a crash, a forced quit or a quit that keeps documents loses nothing.
//
// A copy is one file of UTF-8 text, with the document's line endings, and an
// entry in index.json that says which document it belongs to: its file, its
// unstyled number, its encoding, and the size and modification time the file
// had when the copy was made. Saving a document, closing it or discarding its
// changes removes the copy.
package draft

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/helloxz/zpad/internal/atomicfile"
)

const indexName = "index.json"

// Entry describes one recovery copy.
type Entry struct {
	ID string `json:"id"`
	// Path is the document's file; "" for an untitled document, whose
	// Untitled number names it.
	Path     string `json:"path,omitempty"`
	Untitled int    `json:"untitled,omitempty"`
	// Encoding is what a save of the document writes; the copy itself is
	// always UTF-8.
	Encoding string `json:"encoding,omitempty"`
	// Language is the language an untitled document was given in the status
	// bar, by name; "" for none. A document with a path takes its language
	// from the file name instead.
	Language string `json:"language,omitempty"`
	// Size and ModTime are the file's size and modification time when the
	// copy was made, in Unix nanoseconds; zero for an untitled document. A
	// file that changed since is in conflict with the copy.
	Size    int64 `json:"size,omitempty"`
	ModTime int64 `json:"modTime,omitempty"`
	// SavedAt is when the copy was written, in Unix nanoseconds.
	SavedAt int64 `json:"savedAt,omitempty"`
}

// Store is the directory of recovery copies. Its methods are safe from any
// goroutine.
type Store struct {
	dir     string
	mu      sync.Mutex
	entries map[string]Entry
}

// Open reads the copies in dir, creating it when it is not there. An index
// that cannot be read is kept as index.json.bak and its copies are left on
// disk, for a person to recover.
func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	s := &Store{dir: dir, entries: map[string]Entry{}}
	path := filepath.Join(dir, indexName)
	data, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return s, nil
	case err != nil:
		return nil, err
	}
	var list []Entry
	if err := json.Unmarshal(data, &list); err != nil {
		_ = os.WriteFile(path+".bak", data, 0o600)
		return s, nil
	}
	for _, e := range list {
		if !ValidID(e.ID) {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, e.ID)); err != nil {
			continue
		}
		s.entries[e.ID] = e
	}
	return s, nil
}

// Entries returns the copies, oldest first.
func (s *Store) Entries() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		list = append(list, e)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].SavedAt != list[j].SavedAt {
			return list[i].SavedAt < list[j].SavedAt
		}
		return list[i].ID < list[j].ID
	})
	return list
}

// Text reads the text of the copy id.
func (s *Store) Text(id string) (string, error) {
	if !ValidID(id) {
		return "", errors.New("draft: invalid id")
	}
	data, err := os.ReadFile(filepath.Join(s.dir, id))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Save writes a copy of a document, if alive reports that the copy is still
// the one to write at the moment it would be written. It reports whether the
// copy was written.
func (s *Store) Save(e Entry, write func(io.Writer) error, alive func() bool) (bool, error) {
	if !ValidID(e.ID) {
		return false, errors.New("draft: invalid id")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if alive != nil && !alive() {
		return false, nil
	}
	if e.SavedAt == 0 {
		e.SavedAt = time.Now().UnixNano()
	}
	if err := atomicfile.WriteTo(s.dir, e.ID, 0o600, write); err != nil {
		return false, err
	}
	previous, had := s.entries[e.ID]
	s.entries[e.ID] = e
	if err := s.writeIndexLocked(); err != nil {
		if had {
			s.entries[e.ID] = previous
		} else {
			delete(s.entries, e.ID)
		}
		return false, err
	}
	return true, nil
}

// Remove deletes the copy id, if it is there.
func (s *Store) Remove(id string) error {
	if !ValidID(id) {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.entries[id]; !ok {
		_ = os.Remove(filepath.Join(s.dir, id))
		return nil
	}
	delete(s.entries, id)
	if err := s.writeIndexLocked(); err != nil {
		return err
	}
	err := os.Remove(filepath.Join(s.dir, id))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// SetPath moves a copy's document to path, keeping everything else. It is
// what a rename of the file, or of a folder it is in, asks for.
func (s *Store) SetPath(id, path string) error {
	if !ValidID(id) {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[id]
	if !ok || e.Path == path {
		return nil
	}
	e.Path = path
	s.entries[id] = e
	return s.writeIndexLocked()
}

// Purge deletes every copy, as turning off the keeping of unsaved content
// does.
func (s *Store) Purge() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.RemoveAll(s.dir); err != nil {
		return err
	}
	s.entries = map[string]Entry{}
	return os.MkdirAll(s.dir, 0o700)
}

func (s *Store) writeIndexLocked() error {
	list := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		list = append(list, e)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].SavedAt != list[j].SavedAt {
			return list[i].SavedAt < list[j].SavedAt
		}
		return list[i].ID < list[j].ID
	})
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return atomicfile.Write(s.dir, indexName, data)
}

// ValidID reports whether id can name a copy's file.
func ValidID(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	for _, r := range id {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

// NewID returns an id for a copy: random, so that it does not follow a path
// that may be long, hold a colon on Windows or be written in a case the
// file system folds.
func NewID() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("d%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
