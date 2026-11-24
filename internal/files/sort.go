package files

import (
	"strings"
	"unicode/utf8"
)

// compareNames orders file and directory names the way the sidebar shows
// them: digits compare by value ("2" before "10"), letters compare without
// case, and names that differ only by case put the lowercase spelling first.
// It mirrors the default sorting of Zed's project panel.
func compareNames(a, b string) int {
	if a == b {
		return 0
	}
	ai, bi := 0, 0
	for ai < len(a) && bi < len(b) {
		if isDigit(a[ai]) && isDigit(b[bi]) {
			an, bn := digitRun(a[ai:]), digitRun(b[bi:])
			if c := compareDigitRuns(a[ai:ai+an], b[bi:bi+bn]); c != 0 {
				return c
			}
			ai += an
			bi += bn
			continue
		}
		ra, sizeA := utf8.DecodeRuneInString(a[ai:])
		rb, sizeB := utf8.DecodeRuneInString(b[bi:])
		if ka, kb := foldASCII(ra), foldASCII(rb); ka != kb {
			if ka < kb {
				return -1
			}
			return 1
		}
		ai += sizeA
		bi += sizeB
	}
	if ai < len(a) {
		return 1
	}
	if bi < len(b) {
		return -1
	}
	// The names match apart from case; put the lowercase spelling first so the
	// order never depends on the order the directory was read.
	return strings.Compare(b, a)
}

func isDigit(c byte) bool { return '0' <= c && c <= '9' }

// digitRun counts the leading ASCII digits of s.
func digitRun(s string) int {
	n := 0
	for n < len(s) && isDigit(s[n]) {
		n++
	}
	return n
}

// compareDigitRuns compares two runs of ASCII digits by value. Equal values
// sort the longer run last, so "2" comes before "02".
func compareDigitRuns(a, b string) int {
	na := strings.TrimLeft(a, "0")
	nb := strings.TrimLeft(b, "0")
	if len(na) != len(nb) {
		if len(na) < len(nb) {
			return -1
		}
		return 1
	}
	if c := strings.Compare(na, nb); c != 0 {
		return c
	}
	switch {
	case len(a) < len(b):
		return -1
	case len(a) > len(b):
		return 1
	}
	return 0
}

// foldASCII lowercases ASCII letters, leaving every other rune alone.
func foldASCII(r rune) rune {
	if 'A' <= r && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}
