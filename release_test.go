package main

import (
	"encoding/json"
	"image/png"
	"os"
	"testing"
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
		for _, ext := range a.Ext {
			found[ext] = true
		}
	}
	for _, ext := range []string{"txt", "md", "html", "css", "json", "js"} {
		if !found[ext] {
			t.Fatalf("file association %q missing", ext)
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
