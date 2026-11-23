package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/helloxz/zpad/internal/app"
)

func TestFileAssociations(t *testing.T) {
	data, err := os.ReadFile("mygo.json")
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		FileAssociations []struct {
			Ext      []string `json:"ext"`
			Name     string   `json:"name"`
			MIMEType string   `json:"mimeType"`
		} `json:"fileAssociations"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	if len(cfg.FileAssociations) == 0 {
		t.Fatal("no file associations")
	}
	found := map[string]bool{}
	for _, a := range cfg.FileAssociations {
		if a.Name == "" || a.MIMEType == "" {
			t.Fatalf("file association %v has no name or MIME type", a.Ext)
		}
		for _, ext := range a.Ext {
			if found[ext] {
				// Two entries would write one ProgID twice, the later
				// name winning.
				t.Fatalf("file association %q listed twice", ext)
			}
			found[ext] = true
		}
	}
	for _, ext := range []string{
		// Documents.
		"txt", "md", "markdown", "html", "htm", "css", "json", "xml", "yaml", "toml", "log", "csv",
		// Source code.
		"js", "ts", "tsx", "go", "py", "rs", "c", "cpp", "h", "java", "kt", "swift", "cs",
		"rb", "php", "lua", "sh", "sql", "vue", "svelte", "scss", "less", "diff",
	} {
		if !found[ext] {
			t.Fatalf("file association %q missing", ext)
		}
	}
}

// The open dialog offers the text files the app is associated with: a file
// opened from the file manager must be one the dialog can offer too.
func TestTextExtensionsMatchAssociations(t *testing.T) {
	data, err := os.ReadFile("mygo.json")
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		FileAssociations []struct {
			Ext []string `json:"ext"`
		} `json:"fileAssociations"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	associated := map[string]bool{}
	for _, a := range cfg.FileAssociations {
		for _, ext := range a.Ext {
			associated[ext] = true
		}
	}
	offered := app.TextExtensions()
	if len(offered) != len(associated) {
		t.Fatalf("the open dialog offers %d extensions, mygo.json associates %d", len(offered), len(associated))
	}
	for _, ext := range offered {
		if !associated[ext] {
			t.Fatalf("the open dialog offers %q, which mygo.json does not associate", ext)
		}
	}
}

// mygo.json is the app's only version: mygo build links it into the binary,
// and names the artifacts and the tag with it.
func TestVersionIsAReleaseVersion(t *testing.T) {
	data, err := os.ReadFile("mygo.json")
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	if !releaseVersion.MatchString(cfg.Version) {
		t.Fatalf("mygo.json says version %q, want MAJOR.MINOR.PATCH", cfg.Version)
	}
}

// releaseVersion is what the tag of a release names, as v1.2.3.
var releaseVersion = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// Installed apps accept only updates signed with the private half of the
// key here, and find their manifests in the latest release of updates.github:
// both are compiled into the app by mygo build.
func TestUpdatesAreConfigured(t *testing.T) {
	data, err := os.ReadFile("mygo.json")
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Updates *struct {
			PublicKey string `json:"publicKey"`
			GitHub    string `json:"github"`
		} `json:"updates"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Updates == nil {
		t.Fatal("mygo.json has no updates: the app cannot update itself")
	}
	key, err := base64.StdEncoding.DecodeString(cfg.Updates.PublicKey)
	if err != nil {
		t.Fatalf("updates.publicKey is not base64: %v", err)
	}
	if len(key) != ed25519.PublicKeySize {
		t.Fatalf("updates.publicKey is %d bytes, want an Ed25519 key of %d", len(key), ed25519.PublicKeySize)
	}
	if cfg.Updates.GitHub != "helloxz/zpad" {
		t.Fatalf("updates.github is %q, want helloxz/zpad", cfg.Updates.GitHub)
	}
}

// The update window shows the version's section of CHANGELOG.md, which mygo
// build writes into the manifests: a version without one fails the build.
func TestChangelogHasTheReleaseNotes(t *testing.T) {
	data, err := os.ReadFile("mygo.json")
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	notes, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(notes), "## "+cfg.Version) {
		t.Fatalf("CHANGELOG.md has no section for %s", cfg.Version)
	}
}

// The site's version comes from website/data/release.json, which the website
// workflow rewrites when a release is published. A version typed into the copy
// goes stale on the next release: the hero badge still said v0.3.2 after 0.4.0
// shipped. Templates build the badge from the data, and nothing under website/
// names a version by hand.
func TestWebsiteHasNoHardcodedVersion(t *testing.T) {
	// The icon SVGs are full of numbers like 1.02.08, so a template only counts
	// as naming a version when it carries the v prefix the site uses.
	templates := regexp.MustCompile(`\bv\d+\.\d+\.\d+`)
	copyText := regexp.MustCompile(`\d+\.\d+\.\d+`)

	var files []string
	for _, glob := range []string{"website/content/*.md", "website/i18n/*.toml"} {
		matched, err := filepath.Glob(glob)
		if err != nil {
			t.Fatal(err)
		}
		if len(matched) == 0 {
			t.Fatalf("%s matches nothing: is the site still there?", glob)
		}
		files = append(files, matched...)
	}
	err := filepath.WalkDir("website/layouts", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".html") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		pattern := copyText
		if strings.HasSuffix(path, ".html") {
			pattern = templates
		}
		for i, line := range strings.Split(string(data), "\n") {
			if m := pattern.FindString(line); m != "" {
				t.Errorf("%s:%d names the version %q; derive it from website/data/release.json", path, i+1, m)
			}
		}
	}
}

func TestAppIconSize(t *testing.T) {
	f, err := os.Open("resources/icon.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	b := img.Bounds()
	if b.Dx() != 1024 || b.Dy() != 1024 {
		t.Fatalf("icon is %dx%d, want 1024x1024", b.Dx(), b.Dy())
	}
}
