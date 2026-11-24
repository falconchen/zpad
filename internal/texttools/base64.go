package texttools

import (
	"encoding/base64"
	"strings"
	"unicode/utf8"
)

// EncodeBase64 returns s as standard, padded Base64 on one line.
func EncodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// DecodeBase64 decodes Base64 text, ignoring spaces and line breaks, as
// people copy it from logs and URLs: it takes the standard and the URL-safe
// alphabet, padded or not. The decoded bytes must be UTF-8 text, since the
// editor's buffer holds nothing else.
func DecodeBase64(s string) (string, error) {
	clean := strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\r', '\n', '\v', '\f':
			return -1
		}
		return r
	}, s)
	if clean == "" {
		return "", ErrEmpty
	}

	var (
		decoded []byte
		err     error
	)
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		decoded, err = enc.DecodeString(clean)
		if err == nil {
			break
		}
	}
	if err != nil {
		return "", err
	}

	out := string(decoded)
	if !utf8.ValidString(out) {
		return "", ErrNotText
	}
	return out, nil
}
