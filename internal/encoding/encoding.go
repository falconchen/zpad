// Package encoding converts text files between UTF-8 and the encodings
// ZPad edits. The editor itself works on UTF-8.
package encoding

import (
	"bytes"
	"errors"
	"fmt"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Kind is a text encoding.
type Kind string

const (
	UTF8    Kind = "utf-8"
	UTF8BOM Kind = "utf-8-bom"
	UTF16LE Kind = "utf-16le"
	UTF16BE Kind = "utf-16be"
	GB18030 Kind = "gb18030"
)

// All lists the encodings the status bar offers.
var All = []Kind{UTF8, UTF8BOM, UTF16LE, UTF16BE, GB18030}

// Label is the name shown in the UI.
func Label(k Kind) string {
	switch k {
	case UTF8:
		return "UTF-8"
	case UTF8BOM:
		return "UTF-8 BOM"
	case UTF16LE:
		return "UTF-16 LE"
	case UTF16BE:
		return "UTF-16 BE"
	case GB18030:
		return "GB18030"
	}
	return string(k)
}

// Parse turns a saved setting into a Kind, defaulting to UTF-8.
func Parse(s string) Kind {
	for _, k := range All {
		if string(k) == s {
			return k
		}
	}
	return UTF8
}

var (
	bomUTF8    = []byte{0xEF, 0xBB, 0xBF}
	bomUTF16LE = []byte{0xFF, 0xFE}
	bomUTF16BE = []byte{0xFE, 0xFF}
)

// Detect reads the encoding of data from its BOM, its UTF-16 byte pattern
// or its UTF-8 validity, falling back to GB18030.
func Detect(data []byte) (Kind, error) {
	switch {
	case bytes.HasPrefix(data, bomUTF8):
		return UTF8BOM, nil
	case bytes.HasPrefix(data, bomUTF16LE):
		return UTF16LE, nil
	case bytes.HasPrefix(data, bomUTF16BE):
		return UTF16BE, nil
	}
	if k, ok := detectUTF16(data); ok {
		return k, nil
	}
	if utf8.Valid(data) {
		return UTF8, nil
	}
	// GB18030 covers all Unicode code points and is the fallback for
	// Chinese legacy files. If it cannot decode the bytes, let the user
	// pick another encoding from the status bar.
	if _, err := decodeGB18030(data); err == nil {
		return GB18030, nil
	}
	return GB18030, errors.New("无法识别文件编码，请在状态栏选择编码重新打开")
}

func detectUTF16(data []byte) (Kind, bool) {
	if len(data) < 4 {
		return "", false
	}
	n := len(data)
	if n > 4096 {
		n = 4096
	}
	var evenNul, oddNul int
	for i := 0; i < n; i++ {
		if data[i] != 0 {
			continue
		}
		if i%2 == 0 {
			evenNul++
		} else {
			oddNul++
		}
	}
	pairs := n / 2
	if pairs == 0 {
		return "", false
	}
	// UTF-16LE ASCII has NULs at odd offsets; UTF-16BE at even ones.
	if oddNul > pairs/3 && evenNul < pairs/10 {
		return UTF16LE, true
	}
	if evenNul > pairs/3 && oddNul < pairs/10 {
		return UTF16BE, true
	}
	return "", false
}

// Decode returns data as UTF-8 text.
func Decode(data []byte, kind Kind) (string, error) {
	switch kind {
	case UTF8, UTF8BOM:
		if bytes.HasPrefix(data, bomUTF8) {
			data = data[len(bomUTF8):]
		}
		if !utf8.Valid(data) {
			return "", errors.New("文件不是有效的 UTF-8")
		}
		return string(data), nil
	case UTF16LE:
		return decodeUTF16(data, unicode.LittleEndian)
	case UTF16BE:
		return decodeUTF16(data, unicode.BigEndian)
	case GB18030:
		return decodeGB18030(data)
	default:
		return "", fmt.Errorf("不支持的编码 %q", kind)
	}
}

func decodeUTF16(data []byte, order unicode.Endianness) (string, error) {
	dec := unicode.UTF16(order, unicode.UseBOM).NewDecoder()
	out, _, err := transform.Bytes(dec, data)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(out) {
		return "", errors.New("文件不是有效的 UTF-16")
	}
	return string(out), nil
}

func decodeGB18030(data []byte) (string, error) {
	dec := simplifiedchinese.GB18030.NewDecoder()
	out, _, err := transform.Bytes(dec, data)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(out) {
		return "", errors.New("文件不是有效的 GB18030")
	}
	return string(out), nil
}

// Encode returns text in kind.
func Encode(text string, kind Kind) ([]byte, error) {
	switch kind {
	case UTF8:
		return []byte(text), nil
	case UTF8BOM:
		return append(append([]byte{}, bomUTF8...), []byte(text)...), nil
	case UTF16LE:
		return encodeUTF16(text, unicode.LittleEndian)
	case UTF16BE:
		return encodeUTF16(text, unicode.BigEndian)
	case GB18030:
		enc := simplifiedchinese.GB18030.NewEncoder()
		out, _, err := transform.Bytes(enc, []byte(text))
		if err != nil {
			return nil, err
		}
		return out, nil
	default:
		return nil, fmt.Errorf("不支持的编码 %q", kind)
	}
}

func encodeUTF16(text string, order unicode.Endianness) ([]byte, error) {
	// UseBOM writes a byte order mark, as UTF-16 files usually have.
	enc := unicode.UTF16(order, unicode.UseBOM).NewEncoder()
	out, _, err := transform.Bytes(enc, []byte(text))
	if err != nil {
		return nil, err
	}
	return out, nil
}
