package app

import (
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/egoist/mygo/ui/editor"
)

// untitledLanguages are the languages the status bar offers an untitled
// document: the ones whose first extension the app saves as text, so that
// the file keeps the highlight it is given. One named only by a file name,
// as Makefile and Dockerfile are, cannot name the file its save starts from,
// and is left out. They are ordered by name, so that a list of some fourty
// can be found by reading down it.
var untitledLanguages = sync.OnceValue(func() []*editor.Language {
	all := editor.Languages()
	out := make([]*editor.Language, 0, len(all))
	for _, l := range all {
		if len(l.Extensions) == 0 {
			continue
		}
		if !slices.Contains(textExtensions, strings.TrimPrefix(l.Extensions[0], ".")) {
			continue
		}
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
})
