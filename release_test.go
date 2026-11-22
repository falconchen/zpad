package main

import (
	"encoding/json"
	"image/png"
	"os"
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

func TestVersionMatchesConfig(t *testing.T) {
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
	if cfg.Version != app.Version {
		t.Fatalf("mygo.json says version %q, the About dialog shows %q", cfg.Version, app.Version)
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
