package app

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// keepLogger restores the standard logger a test redirects.
func keepLogger(t *testing.T) {
	t.Helper()
	out, flags, prefix := log.Writer(), log.Flags(), log.Prefix()
	t.Cleanup(func() {
		log.SetOutput(out)
		log.SetFlags(flags)
		log.SetPrefix(prefix)
	})
}

func TestSetupLoggingWritesTheFile(t *testing.T) {
	keepLogger(t)
	t.Setenv("ZPAD_LOG", "")
	dir := t.TempDir()
	path := SetupLogging(dir)
	if path != filepath.Join(dir, LogName) {
		t.Fatalf("log path %q, want %q", path, filepath.Join(dir, LogName))
	}
	log.Print("hello from the test")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "hello from the test") {
		t.Fatalf("log does not hold what was written: %q", data)
	}
}

func TestSetupLoggingRotatesTheOldLog(t *testing.T) {
	keepLogger(t)
	t.Setenv("ZPAD_LOG", "")
	dir := t.TempDir()
	path := filepath.Join(dir, LogName)
	if err := os.WriteFile(path, bytes.Repeat([]byte("x"), logMax+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := SetupLogging(dir); got != path {
		t.Fatalf("log path %q, want %q", got, path)
	}
	if info, err := os.Stat(path + ".1"); err != nil || info.Size() <= logMax {
		t.Fatalf("the old log was not moved aside: %v", err)
	}
	if info, err := os.Stat(path); err != nil || info.Size() != 0 {
		t.Fatalf("a fresh log was not started: %v", err)
	}
}

func TestSetupLoggingOff(t *testing.T) {
	keepLogger(t)
	t.Setenv("ZPAD_LOG", "off")
	if path := SetupLogging(t.TempDir()); path != "" {
		t.Fatalf("logging off returned the path %q", path)
	}
}

// TestSlowFrameSummary makes a summary due and checks what it says, without
// waiting the ten seconds one waits in a running app.
func TestSlowFrameSummary(t *testing.T) {
	keepLogger(t)
	var buf bytes.Buffer
	log.SetOutput(&buf)
	frames.reset()
	t.Cleanup(frames.reset)

	for range frameWarm + 1 {
		NoteFrame(time.Millisecond) // opening frames, which say nothing
	}
	frames.mu.Lock()
	frames.since = time.Now().Add(-frameReport - time.Second)
	frames.mu.Unlock()
	NoteFrame(30 * time.Millisecond)

	out := buf.String()
	if !strings.Contains(out, "frames 10s:") || !strings.Contains(out, "slow=1") {
		t.Fatalf("no summary of a slow frame: %q", out)
	}
	if !strings.Contains(out, "max=30.0ms") {
		t.Fatalf("summary misses the slow frame: %q", out)
	}
}

// TestFastFramesStayQuiet is the promise that the log is not written to
// while the app runs well.
func TestFastFramesStayQuiet(t *testing.T) {
	keepLogger(t)
	var buf bytes.Buffer
	log.SetOutput(&buf)
	frames.reset()
	t.Cleanup(frames.reset)

	for range 500 {
		NoteFrame(2 * time.Millisecond)
	}
	if buf.Len() != 0 {
		t.Fatalf("fast frames were reported: %q", buf.String())
	}
}
