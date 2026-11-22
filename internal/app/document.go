package app

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo/ui/editor"
	"github.com/helloxz/zpad/internal/atomicfile"
	"github.com/helloxz/zpad/internal/encoding"
)

const maxFileSize = 20 << 20

// Document is an open file and the text the editor edits.
type Document struct {
	ID       string
	Path     string
	Buffer   *editor.Buffer
	Encoding encoding.Kind
	Raw      []byte
	Size     int64
}

// loadDocument reads a file and detects its encoding.
func loadDocument(path string) (*Document, error) {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, errors.New("这是一个文件夹")
	}
	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("文件超过 %d MB，M1 暂不支持", maxFileSize>>20)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	kind, err := encoding.Detect(data)
	if err != nil {
		return nil, err
	}
	if kind != encoding.UTF16LE && kind != encoding.UTF16BE && isBinary(data) {
		return nil, errors.New("这是二进制文件，无法作为文本编辑")
	}
	return newDocument(path, data, kind, info.Size())
}

// loadDocumentWith reads a file again, interpreting its bytes as kind.
func loadDocumentWith(path string, kind encoding.Kind) (*Document, error) {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, errors.New("这是一个文件夹")
	}
	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("文件超过 %d MB，M1 暂不支持", maxFileSize>>20)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if kind != encoding.UTF16LE && kind != encoding.UTF16BE && isBinary(data) {
		return nil, errors.New("这是二进制文件，无法作为文本编辑")
	}
	return newDocument(path, data, kind, info.Size())
}

func newDocument(path string, data []byte, kind encoding.Kind, size int64) (*Document, error) {
	text, err := encoding.Decode(data, kind)
	if err != nil {
		return nil, err
	}
	buf := editor.NewBuffer(text)
	buf.SetLanguage(editor.LanguageFor(path))
	return &Document{
		ID:       path,
		Path:     path,
		Buffer:   buf,
		Encoding: kind,
		Raw:      data,
		Size:     size,
	}, nil
}

func newUntitled(id string, defaultEncoding string) *Document {
	buf := editor.NewBuffer("")
	return &Document{
		ID:       id,
		Buffer:   buf,
		Encoding: encoding.Parse(defaultEncoding),
	}
}

func (d *Document) Name() string {
	if d.Path == "" {
		return "未命名"
	}
	return filepath.Base(d.Path)
}

// DisplayPath is the path relative to root when it is inside it.
func (d *Document) DisplayPath(root string) string {
	if d.Path == "" {
		return "未命名"
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
		return nil, err
	}
	back, err := encoding.Decode(data, d.Encoding)
	if err != nil || back != text {
		return nil, errors.New("目标编码无法完整表示当前文本，请改用 UTF-8 保存")
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
