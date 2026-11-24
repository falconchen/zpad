package texttools

import (
	"bytes"
	"encoding/json"
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"strings"
)

// FormatJSON indents a JSON value with two spaces, keeping its member order
// and the exact spelling of its numbers, strings and escapes. Whitespace in
// the input is the only thing that changes.
//
// The output is bounded: nesting deeper than MaxJSONDepth is refused, and
// the formatted text is streamed into a buffer that stops at MaxJSONOutput.
// Indentation grows with depth, so a small crafted value can ask for
// gigabytes otherwise.
func FormatJSON(s string) (string, error) {
	data := []byte(s)
	if !json.Valid(data) {
		return "", ErrJSON
	}
	if depth := jsonDepth(data); depth > MaxJSONDepth {
		return "", fmt.Errorf("%w: depth %d", ErrJSONDepth, depth)
	}

	out := &limitedBuffer{limit: MaxJSONOutput}
	// Duplicate names are kept, as Compact and encoding/json keep them: a
	// formatting tool should not refuse text the editor accepted.
	opts := jsontext.AllowDuplicateNames(true)
	dec := jsontext.NewDecoder(strings.NewReader(s), opts)
	enc := jsontext.NewEncoder(out, jsontext.WithIndent("  "), opts)
	v, err := dec.ReadValue()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrJSON, err)
	}
	if err := enc.WriteValue(v); err != nil {
		if errors.Is(err, ErrJSONOutput) {
			return "", ErrJSONOutput
		}
		return "", fmt.Errorf("%w: %v", ErrJSON, err)
	}
	// The encoder ends a top-level value with a newline; the selection
	// should not gain one.
	return strings.TrimSuffix(out.String(), "\n"), nil
}

// MinifyJSON removes every space, tab and line break that is not part of a
// string. The result never grows, so it needs no output bound.
func MinifyJSON(s string) (string, error) {
	data := []byte(s)
	if !json.Valid(data) {
		return "", ErrJSON
	}
	var out bytes.Buffer
	if err := json.Compact(&out, data); err != nil {
		return "", fmt.Errorf("%w: %v", ErrJSON, err)
	}
	return out.String(), nil
}

// jsonDepth returns the deepest object or array nesting of a valid JSON
// value.
func jsonDepth(b []byte) int {
	depth, deepest := 0, 0
	inString, escape := false, false
	for _, c := range b {
		if inString {
			switch {
			case escape:
				escape = false
			case c == '\\':
				escape = true
			case c == '"':
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{', '[':
			depth++
			if depth > deepest {
				deepest = depth
			}
		case '}', ']':
			depth--
		}
	}
	return deepest
}

// limitedBuffer holds what is written until it would exceed limit.
type limitedBuffer struct {
	strings.Builder
	limit int
}

func (w *limitedBuffer) Write(p []byte) (int, error) {
	if w.Builder.Len()+len(p) > w.limit {
		return 0, ErrJSONOutput
	}
	return w.Builder.Write(p)
}
