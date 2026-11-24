package app

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
)

// The site's screenshots are rendered in memory here instead of being taken
// from a window: no screen recording permission, the same demo workspace every
// time, and one set per language and theme.
//
// The gallery keeps the default pair, except the two newer scenes, which show
// named themes: the Markdown preview in GitHub Light and the search in One
// Dark. The hero is the split scene in those same two named themes, so the
// site can show GitHub Light and swap to One Dark as the pointer rests on it.
//
//	go run ./... is not how this runs; set the output and run the one test:
//
//	ZPAD_SHOTS=/tmp/shots go test ./internal/app -run TestSiteScreenshots
func TestSiteScreenshots(t *testing.T) {
	out := os.Getenv("ZPAD_SHOTS")
	if out == "" {
		t.Skip("set ZPAD_SHOTS=<dir> to write the site's screenshots")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}

	scenes := []siteScene{
		{name: "split", set: splitScene},
		{name: "split", dark: true, set: splitScene},
		{name: "settings", set: settingsScene, drive: openThemeMenu},
		{name: "settings", dark: true, set: settingsScene, drive: openThemeMenu},
		{name: "preview", theme: siteLightTheme, set: previewScene},
		{name: "search", dark: true, theme: siteDarkTheme, set: searchScene},
		{name: "hero", theme: siteLightTheme, set: splitScene},
		{name: "hero", dark: true, theme: siteDarkTheme, set: splitScene},
	}

	for _, l := range []struct {
		tag, suffix, notes string
	}{
		{"zh-Hans", "", siteNotesZH},
		{"en", "-en", siteNotesEN},
	} {
		tag := l.tag
		i18n.SetLocaleFunc(func() string { return tag })
		dir := filepath.Join(t.TempDir(), "zpad-demo")
		writeSiteDemo(t, dir, l.notes)

		for _, scene := range scenes {
			name := scene.name + l.suffix
			if scene.dark {
				name += "-dark"
			}
			img := renderSiteApp(t, dir, scene)
			f, err := os.Create(filepath.Join(out, name+".png"))
			if err != nil {
				t.Fatal(err)
			}
			if err := png.Encode(f, img); err != nil {
				f.Close()
				t.Fatal(err)
			}
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
}

// Named themes the site shows off: the gallery's two newer scenes and both
// halves of the hero's hover.
const (
	siteLightTheme = "github-light"
	siteDarkTheme  = "one-dark"
)

// siteScene is one screenshot: the scene that fills the window, the theme it
// draws with, and an optional step after the first frame, as a click that
// opens a drop-down for the shot.
type siteScene struct {
	name  string
	dark  bool
	theme string // a named theme; "" draws the default ZPad light or dark
	set   func(*testing.T, *App)
	drive func(*testing.T, *ui.Tester, *App)
}

func renderSiteApp(t *testing.T, dir string, scene siteScene) image.Image {
	t.Helper()
	s := settings.Default()
	s.SidebarOpen = true
	if scene.theme != "" {
		s.Theme = scene.theme
	}
	a := New(s, t.TempDir())
	if scene.theme == "" {
		a.Dark = scene.dark
	}
	a.setWorkspace(dir)
	scene.set(t, a)

	tt := ui.NewTester(a.View, 1400, 900)
	tt.SetScale(2)
	tt.SetDark(scene.dark)
	tt.SetWindowActive(true)
	tt.Frame()
	if scene.drive != nil {
		scene.drive(t, tt, a)
		tt.Frame()
	}
	return tt.Image()
}

func siteDoc(t *testing.T, a *App, name string) *Document {
	t.Helper()
	doc, err := loadDocument(filepath.Join(a.root, name), 20)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func splitScene(t *testing.T, a *App) {
	a.addDocument(siteDoc(t, a, "notes.md"))
	a.toggleSplit()
	a.addDocumentAt(1, siteDoc(t, a, "main.go"))
}

// settingsScene opens the settings sheet; openThemeMenu then drops its theme
// list down, so the shot carries the named themes as well.
func settingsScene(t *testing.T, a *App) {
	splitScene(t, a)
	a.openSettings()
}

func openThemeMenu(t *testing.T, tt *ui.Tester, a *App) {
	t.Helper()
	name, _ := themeOptionParts(themeLabel(a.Settings.Theme))
	if err := tt.Click(name); err != nil {
		t.Fatalf("open the theme drop-down: %v", err)
	}
}

func previewScene(t *testing.T, a *App) {
	a.addDocument(siteDoc(t, a, "notes.md"))
	a.ToggleMarkdownPreview()
}

// searchScene opens every demo file, so the panel carries the matches of the
// open tabs rather than an almost empty list.
func searchScene(t *testing.T, a *App) {
	for _, name := range []string{"main.go", "style.css", "index.html", "data.json", "notes.md"} {
		a.addDocument(siteDoc(t, a, name))
	}
	a.searchOpen = true
	a.searchQuery = "zpad"
	a.startSearch()
}

const siteNotesZH = `# 发布清单

ZPad 0.5.0 的发布准备。

## 待办

- [x] 打包 dmg 与安装包
- [ ] 更新官网截图
- [ ] 检查自动更新

## 编码

保存时保留编码和行尾，UTF-8 与 GB18030 都能打开。

> 只做编辑器该做的事。

` + "```go" + `
cfg := config{Name: "zpad", Size: 42}
` + "```" + `

| 平台 | 包 |
| --- | --- |
| macOS | dmg |
| Windows | exe |
| Linux | deb |
`

const siteNotesEN = `# Release checklist

Preparing ZPad 0.5.0 for release.

## To do

- [x] Build the dmg and the installer
- [ ] Refresh the site's screenshots
- [ ] Check the updater

## Encodings

Encoding and line endings survive a save,
from UTF-8 to GB18030.

> It does the editing, and only that.

` + "```go" + `
cfg := config{Name: "zpad", Size: 42}
` + "```" + `

| Platform | Package |
| --- | --- |
| macOS | dmg |
| Windows | exe |
| Linux | deb |
`

func writeSiteDemo(t *testing.T, dir, notes string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"notes.md": notes,
		"main.go": `package main

import (
	"fmt"
	"os"
)

type config struct {
	Name string
	Size int
}

func main() {
	cfg := config{Name: "zpad", Size: 42}
	fmt.Printf("hello, %s (%d)\n", cfg.Name, cfg.Size)
	os.Exit(0)
}
`,
		"style.css": `:root {
  --bg: #ffffff;
  --fg: #111827;
  --accent: #4f46e5;
}

body {
  margin: 0;
  background: var(--bg);
  color: var(--fg);
  font: 16px/1.6 system-ui;
}
`,
		"index.html": `<!doctype html>
<html lang="zh-Hans">
  <head>
    <meta charset="utf-8">
    <title>ZPad</title>
  </head>
  <body>
    <h1>ZPad</h1>
    <p>轻量、原生、够快。</p>
  </body>
</html>
`,
		"data.json": `{
  "name": "zpad",
  "version": "0.5.0",
  "platforms": ["darwin", "windows"]
}
`,
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
