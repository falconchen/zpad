package app

import (
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/egoist/mygo/ui/editor"
)

// commonUntitledLanguages head the status bar's list, in this order: the
// languages a file made from nothing most often is. The status bar shows
// some ten rows at a time, so the rest would be a scroll away.
var commonUntitledLanguages = []string{
	"JSON", "YAML", "Markdown", "HTML", "CSS", "JavaScript",
	"TypeScript", "PHP", "Python", "Shell", "Go",
}

// untitledLanguageGroups are the languages the status bar offers an untitled
// document, in the order its menu shows them: those of
// commonUntitledLanguages first, then the rest by name. A language is
// offered when its first extension is one the app saves as text, so that the
// file keeps the highlight it is given; one named only by a file name, as
// Makefile and Dockerfile are, cannot name the file its save starts from, and
// is left out.
var untitledLanguageGroups = sync.OnceValue(func() languageGroups {
	byName := map[string]*editor.Language{}
	for _, l := range editor.Languages() {
		if len(l.Extensions) == 0 {
			continue
		}
		if !slices.Contains(textExtensions, strings.TrimPrefix(l.Extensions[0], ".")) {
			continue
		}
		byName[l.Name] = l
	}
	var groups languageGroups
	for _, name := range commonUntitledLanguages {
		if l, ok := byName[name]; ok {
			groups.common = append(groups.common, l)
			delete(byName, name)
		}
	}
	for _, l := range byName {
		groups.rest = append(groups.rest, l)
	}
	sort.Slice(groups.rest, func(i, j int) bool {
		return strings.ToLower(groups.rest[i].Name) < strings.ToLower(groups.rest[j].Name)
	})
	return groups
})

// languageGroups are the languages of the status bar's menu, split where the
// menu shows a rule: the common ones, and the rest.
type languageGroups struct {
	common, rest []*editor.Language
}

// all returns both groups in the order the menu shows them.
func (g languageGroups) all() []*editor.Language {
	return append(slices.Clone(g.common), g.rest...)
}
