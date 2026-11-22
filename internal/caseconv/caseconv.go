// Package caseconv converts the case of selected text.
package caseconv

import (
	"strings"
	"unicode"
)

// Mode is a case conversion.
type Mode string

const (
	Upper  Mode = "upper"
	Lower  Mode = "lower"
	Title  Mode = "title"
	Invert Mode = "invert"
)

// Apply converts s with mode.
func Apply(s string, mode Mode) string {
	switch mode {
	case Upper:
		return strings.ToUpper(s)
	case Lower:
		return strings.ToLower(s)
	case Title:
		return title(s)
	case Invert:
		return invert(s)
	}
	return s
}

func title(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	start := true
	for _, r := range s {
		if unicode.IsLetter(r) {
			if start {
				b.WriteRune(unicode.ToUpper(r))
				start = false
			} else {
				b.WriteRune(unicode.ToLower(r))
			}
			continue
		}
		b.WriteRune(r)
		// An apostrophe keeps the word going: "don't" -> "Don't".
		if !unicode.IsDigit(r) && r != '\'' && r != '’' {
			start = true
		}
	}
	return b.String()
}

func invert(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case unicode.IsUpper(r):
			b.WriteRune(unicode.ToLower(r))
		case unicode.IsLower(r):
			b.WriteRune(unicode.ToUpper(r))
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
