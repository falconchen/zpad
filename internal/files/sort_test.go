package files

import (
	"slices"
	"testing"
)

func TestCompareNamesSortsNaturally(t *testing.T) {
	names := []string{
		"10月", "11月", "3月", "4月", "7月", "8月", "9月", "1月", "2月", "12月",
		"file10", "file2", "file1", "file02", "File02",
		"a.txt", "A.txt", "zebra", "ä", "000", "0", "01", "1",
	}
	want := []string{
		"0", "000", "1", "1月", "01", "2月", "3月", "4月", "7月", "8月", "9月", "10月", "11月", "12月",
		"a.txt", "A.txt",
		"file1", "file2", "file02", "File02", "file10",
		"zebra", "ä",
	}
	slices.SortFunc(names, compareNames)
	if !slices.Equal(names, want) {
		t.Fatalf("sorted names\n got %q\nwant %q", names, want)
	}
}

func TestCompareNamesIsAntisymmetric(t *testing.T) {
	names := []string{
		"0", "00", "1", "01", "2", "10", "a", "A", "a1", "A01", "a01", "ä", "Ä",
		"file", "file.", "file1", "file01", "file1a", "file10", "1月", "10月",
	}
	for _, a := range names {
		if got := compareNames(a, a); got != 0 {
			t.Fatalf("compareNames(%q, %q) = %d, want 0", a, a, got)
		}
		for _, b := range names {
			ab, ba := compareNames(a, b), compareNames(b, a)
			if ab != -ba {
				t.Fatalf("compareNames(%q, %q) = %d but compareNames(%q, %q) = %d", a, b, ab, b, a, ba)
			}
		}
	}
}
