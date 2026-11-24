// Package texttools transforms selected text: Base64, MD5 and JSON.
//
// Every function works on one selection, which the application bounds by
// MaxInput bytes before it calls here: a tool copies its input, and JSON
// formatting can write several times the input, so the limit is what keeps
// a whole-file selection from spiking the heap.
package texttools

import "errors"

const (
	// MaxInput is the most selected text a tool works on, counted as UTF-8
	// bytes across all selections of one action.
	MaxInput = 512 << 10
	// MaxJSONOutput bounds the text a formatted JSON may become. The
	// indentation of a deeply nested value grows quadratically, so the cap
	// is enforced while the output streams rather than after.
	MaxJSONOutput = 8 << 20
	// MaxJSONDepth is the deepest nesting a value may have. It keeps the
	// indentation of a crafted value from reaching MaxJSONOutput at all.
	MaxJSONDepth = 128
)

var (
	// ErrEmpty says there was nothing to decode.
	ErrEmpty = errors.New("nothing to decode")
	// ErrNotText says decoded bytes are not UTF-8 text, which the editor's
	// buffer cannot hold.
	ErrNotText = errors.New("decoded data is not UTF-8 text")
	// ErrJSON wraps a JSON syntax error: the value is not valid JSON.
	ErrJSON = errors.New("invalid JSON")
	// ErrJSONDepth says the value nests deeper than MaxJSONDepth.
	ErrJSONDepth = errors.New("JSON is nested too deeply")
	// ErrJSONOutput says formatting the value would exceed MaxJSONOutput.
	ErrJSONOutput = errors.New("formatted JSON is too large")
)
