package app

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/egoist/mygo/ui/editor"
	"github.com/egoist/mygo/ui/markdown"
	"github.com/helloxz/zpad/internal/atomicfile"
	"github.com/helloxz/zpad/internal/encoding"
	"github.com/helloxz/zpad/locales"
)

const maxLines = 500_000

// Document is an open file and the text the editor edits.
type Document struct {
	ID       string
	Path     string
	Buffer   *editor.Buffer
	Encoding encoding.Kind
	// Language is the file's language, shown in the status bar; the buffer
	// gets it only when the file is short enough to highlight.
	Language *editor.Language
	Size     int64
	ModTime  time.Time
	ReadOnly bool
	// UntitledSeq is the number the document was created with, for the name
	// its save dialog starts from; 0 for a document with a path.
	UntitledSeq int

	// md are the Markdown preview's options, shared by every preview of the
	// document; nil until one shows. mdPath is the path their image base
	// came from, and previewText caches the buffer at previewVersion.
	md             *markdown.Options
	mdPath         string
	previewText    string
	previewVersion uint64
	previewLoaded  bool
}

// setLanguage gives the buffer the file's language, or none when the text
// is too long to highlight, and remembers the language for the status bar.
// It runs after the text is in the buffer: both reloads replace the text
// first.
func (d *Document) setLanguage(path string) {
	d.Language = editor.LanguageFor(path)
	if d.Language != nil && d.Buffer.LineCount() > maxLines {
		// A file that long is read-only anyway, and highlighting it would
		// keep the highlighter busy for seconds as it scrolls.
		log.Printf("no highlight: %s %d lines > %d (%s)", filepath.Base(path),
			d.Buffer.LineCount(), maxLines, d.Language.Name)
		d.Buffer.SetLanguage(nil)
		return
	}
	d.Buffer.SetLanguage(d.Language)
}

// loadDocument reads a file and detects its encoding.
func loadDocument(path string, maxMB int) (*Document, error) {
	if maxMB <= 0 {
		maxMB = 20
	}
	maxBytes := int64(maxMB) << 20
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, errors.New(locales.IsFolder())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	kind, err := encoding.Detect(data)
	if err != nil {
		return nil, errors.New(locales.UnknownEncoding())
	}
	if kind != encoding.UTF16LE && kind != encoding.UTF16BE && isBinary(data) {
		return nil, errors.New(locales.BinaryFile())
	}
	doc, err := newDocument(path, data, kind, info.Size())
	if err != nil {
		return nil, err
	}
	doc.ModTime = info.ModTime()
	doc.ReadOnly = info.Size() > maxBytes || doc.Buffer.LineCount() > maxLines
	return doc, nil
}

// loadDocumentWith reads a file again, interpreting its bytes as kind.
func loadDocumentWith(path string, kind encoding.Kind, maxMB int) (*Document, error) {
	if maxMB <= 0 {
		maxMB = 20
	}
	maxBytes := int64(maxMB) << 20
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, errors.New(locales.IsFolder())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if kind != encoding.UTF16LE && kind != encoding.UTF16BE && isBinary(data) {
		return nil, errors.New(locales.BinaryFile())
	}
	doc, err := newDocument(path, data, kind, info.Size())
	if err != nil {
		return nil, err
	}
	doc.ModTime = info.ModTime()
	doc.ReadOnly = info.Size() > maxBytes || doc.Buffer.LineCount() > maxLines
	return doc, nil
}

func newDocument(path string, data []byte, kind encoding.Kind, size int64) (*Document, error) {
	text, err := encoding.Decode(data, kind)
	if err != nil {
		return nil, errors.New(locales.DecodeFailed())
	}
	doc := &Document{
		ID:       path,
		Path:     path,
		Buffer:   editor.NewBuffer(text),
		Encoding: kind,
		Size:     size,
	}
	doc.setLanguage(path)
	return doc, nil
}

func newUntitled(seq int, defaultEncoding string) *Document {
	buf := editor.NewBuffer("")
	return &Document{
		ID:          fmt.Sprintf("untitled-%d", seq),
		Buffer:      buf,
		Encoding:    encoding.Parse(defaultEncoding),
		UntitledSeq: seq,
	}
}

// name is what the tab, the title and the prompts call the document: a saved
// file by its file name, an untitled one by the number it was created with.
// The number stays as the other tabs close, so that a prompt naming one points
// at it rather than at whichever tab moved into its place.
func (d *Document) Name() string {
	if d.Path == "" {
		return d.untitledName()
	}
	return filepath.Base(d.Path)
}

func (d *Document) untitledName() string {
	if d.UntitledSeq <= 0 {
		return locales.Untitled()
	}
	return locales.UntitledN(d.UntitledSeq)
}

// saveName is the file name a save dialog starts with. An untitled
// document gets an English name with its creation number and the text
// extension, so the dialog does not offer a localized name that the user
// would have to replace.
func (d *Document) saveName() string {
	if d.Path != "" {
		return filepath.Base(d.Path)
	}
	seq := d.UntitledSeq
	if seq <= 0 {
		seq = 1
	}
	return fmt.Sprintf("untitled-%d.txt", seq)
}

// DisplayPath is the path relative to root when it is inside it.
func (d *Document) DisplayPath(root string) string {
	if d.Path == "" {
		return d.untitledName()
	}
	if root != "" {
		if rel, err := filepath.Rel(root, d.Path); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return rel
		}
	}
	return d.Path
}

// encoded returns the buffer as bytes in the document's encoding, with the
// line endings the buffer remembers.
func (d *Document) encoded() ([]byte, error) {
	var utf8 bytes.Buffer
	if _, err := d.Buffer.WriteTo(&utf8); err != nil {
		return nil, err
	}
	text := utf8.String()
	data, err := encoding.Encode(text, d.Encoding)
	if err != nil {
		return nil, errors.New(locales.InvalidEncoding())
	}
	back, err := encoding.Decode(data, d.Encoding)
	if err != nil || back != text {
		return nil, errors.New(locales.InvalidEncoding())
	}
	return data, nil
}

// writeDocument writes d to path, keeping the file's mode when it exists.
func writeDocument(d *Document, path string) ([]byte, error) {
	data, err := d.encoded()
	if err != nil {
		return nil, err
	}
	mode := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	if err := atomicfile.WriteFile(path, data, mode); err != nil {
		return nil, err
	}
	return data, nil
}

func isBinary(data []byte) bool {
	n := len(data)
	if n > 8192 {
		n = 8192
	}
	for i := 0; i < n; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}
