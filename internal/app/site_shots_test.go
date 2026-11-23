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

	scenes := []struct {
		name string
		set  func(*testing.T, *App)
	}{
		{"split", splitScene},
		{"settings", settingsScene},
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
			for _, dark := range []bool{false, true} {
				name := scene.name + l.suffix
				if dark {
					name += "-dark"
				}
				img := renderSiteApp(t, dir, dark, scene.set)
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
}

func renderSiteApp(t *testing.T, dir string, dark bool, setup func(*testing.T, *App)) image.Image {
	t.Helper()
	s := settings.Default()
	s.SidebarOpen = true
	a := New(s, t.TempDir())
	a.Dark = dark
	a.setWorkspace(dir)
	setup(t, a)

	tt := ui.NewTester(a.View, 1400, 900)
	tt.SetScale(2)
	tt.SetDark(dark)
	tt.SetWindowActive(true)
	tt.Frame()
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

func settingsScene(t *testing.T, a *App) {
	a.addDocument(siteDoc(t, a, "notes.md"))
	a.toggleSplit()
	a.addDocumentAt(1, siteDoc(t, a, "main.go"))
	a.openSettings()
}

const siteNotesZH = `# 发布清单

- 打包 dmg 与安装包
- 更新官网截图
- 检查自动更新

保存时保留编码和行尾，
UTF-8 与 GB18030 都能打开。
`

const siteNotesEN = `# Release notes

- Build the dmg and the installer
- Refresh the site's screenshots
- Check the updater

Encoding and line endings survive a
save, from UTF-8 to GB18030.
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
  "version": "0.3.2",
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
