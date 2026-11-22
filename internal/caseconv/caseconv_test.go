package caseconv

import "testing"

func TestApply(t *testing.T) {
	cases := []struct {
		in   string
		mode Mode
		want string
	}{
		{"hello world", Upper, "HELLO WORLD"},
		{"Hello World", Lower, "hello world"},
		{"hello world", Title, "Hello World"},
		{"don't stop", Title, "Don't Stop"},
		{"MiXeD 123 abc", Invert, "mIxEd 123 ABC"},
		{"中文 English", Upper, "中文 ENGLISH"},
		{"中文 English", Lower, "中文 english"},
		{"中文 English", Title, "中文 English"},
		{"中文 English", Invert, "中文 eNGLISH"},
	}
	for _, tc := range cases {
		if got := Apply(tc.in, tc.mode); got != tc.want {
			t.Errorf("Apply(%q, %s) = %q, want %q", tc.in, tc.mode, got, tc.want)
		}
	}
}

func TestUnknownModeLeavesText(t *testing.T) {
	if got := Apply("Keep", Mode("nope")); got != "Keep" {
		t.Fatalf("got %q", got)
	}
}
