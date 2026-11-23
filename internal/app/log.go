package app

import (
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

// The diagnostic log, next to the settings: what mygo says of the renderer
// it chose, and how long the app's frames and file operations take. A
// Windows build is a GUI program without a console, so what mygo prints
// there would otherwise be lost.
const (
	LogName = "zpad.log"
	// logMax is how large the log may grow before the next launch moves it
	// aside as zpad.log.1: two files of at most logMax each.
	logMax = 1 << 20
	// frameSlow is the frame time worth telling of: a 60 Hz frame is
	// 16.7 ms, leaving room for the GPU's present.
	frameSlow = 12 * time.Millisecond
	// frameReport is how rarely a run of slow frames is summarized: only
	// when there is a slow frame to tell of.
	frameReport = 10 * time.Second
	// frameWindow is how many recent frames a summary takes its
	// percentiles from.
	frameWindow = 256
	// frameWarm is how many opening frames a summary leaves out, which the
	// first shaping of text makes slow without telling anything.
	frameWarm = 10
)

// SetupLogging sends the standard logger to dir/zpad.log and to stderr, so
// that what mygo logs survives a build without a console, and returns the
// path it opened. It returns "" when logging is off (ZPAD_LOG=off) or the
// file cannot be written: logging never keeps the editor from starting.
// ZPAD_LOG names another file.
func SetupLogging(dir string) string {
	if os.Getenv("ZPAD_LOG") == "off" {
		return ""
	}
	path := filepath.Join(dir, LogName)
	if p := os.Getenv("ZPAD_LOG"); p != "" {
		path = p
	}
	if info, err := os.Stat(path); err == nil && info.Size() > logMax {
		// One older log, so that a long-lived install does not grow a file
		// without bound; the next launch starts a fresh one.
		os.Rename(path, path+".1")
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return ""
	}
	// The file goes first: io.MultiWriter stops at the first error, and
	// stderr may not be writable on Windows.
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(io.MultiWriter(f, os.Stderr))
	return path
}

// frames summarizes the frames that took long to build and paint. note runs
// every frame and allocates nothing; a summary is made only now and then,
// and only when a frame ran slow.
var frames frameStats

// NoteFrame records how long a frame took to build and paint, and writes a
// summary when frames have been running slow. It times the app's own work
// on the main thread, not the GPU's present: a summary points at the
// editor, while a log quiet through a stuttering window points at the
// renderer.
func NoteFrame(d time.Duration) { frames.note(d) }

type frameStats struct {
	mu    sync.Mutex
	ring  [frameWindow]uint32 // the recent frames, in microseconds
	next  int
	total int
	seen  int // frames since the last summary
	slow  int // those of them past frameSlow
	max   uint32
	since time.Time
}

// reset forgets the frames counted so far, for tests.
func (f *frameStats) reset() {
	f.mu.Lock()
	f.next, f.total, f.seen, f.slow, f.max, f.since = 0, 0, 0, 0, 0, time.Time{}
	f.mu.Unlock()
}

func (f *frameStats) note(d time.Duration) {
	us := uint32(min(max(d.Microseconds(), 0), math.MaxUint32))
	f.mu.Lock()
	f.total++
	if f.total <= frameWarm {
		f.mu.Unlock()
		return
	}
	if f.since.IsZero() {
		f.since = time.Now()
	}
	f.ring[f.next] = us
	f.next = (f.next + 1) % frameWindow
	f.seen++
	if d >= frameSlow {
		f.slow++
	}
	if us > f.max {
		f.max = us
	}
	if f.slow == 0 || time.Since(f.since) < frameReport {
		f.mu.Unlock()
		return
	}
	n := min(f.seen, frameWindow)
	var sorted [frameWindow]uint32
	copy(sorted[:], f.ring[:n])
	slices.Sort(sorted[:n])
	p50, p95 := sorted[n/2], sorted[min(n-1, n*95/100)]
	seen, slow, peak := f.seen, f.slow, f.max
	f.seen, f.slow, f.max, f.since = 0, 0, 0, time.Now()
	f.mu.Unlock()
	log.Printf("frames %s: n=%d p50=%.1fms p95=%.1fms max=%.1fms slow=%d",
		frameReport, seen, float64(p50)/1000, float64(p95)/1000, float64(peak)/1000, slow)
}
