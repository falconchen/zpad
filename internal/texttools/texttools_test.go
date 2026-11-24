package texttools

import (
	"errors"
	"strings"
	"testing"
)

func TestEncodeBase64(t *testing.T) {
	cases := map[string]string{
		"":             "",
		"hello":        "aGVsbG8=",
		"中文":           "5Lit5paH",
		"??>":          "Pz8+",
		"a\nb":         "YQpi",
		"line1\nline2": "bGluZTEKbGluZTI=",
	}
	for in, want := range cases {
		if got := EncodeBase64(in); got != want {
			t.Errorf("EncodeBase64(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDecodeBase64(t *testing.T) {
	cases := map[string]string{
		"aGVsbG8=":     "hello",
		"aGVsbG8":      "hello", // no padding
		"5Lit5paH":     "中文",
		"Pz8+":         "??>",
		"Pz8-":         "??>", // URL-safe alphabet
		"Pz8":          "??",  // raw, no padding
		" aGVs\nbG8= ": "hello",
		"YQpi":         "a\nb",
	}
	for in, want := range cases {
		got, err := DecodeBase64(in)
		if err != nil {
			t.Errorf("DecodeBase64(%q): %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("DecodeBase64(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDecodeBase64Empty(t *testing.T) {
	for _, in := range []string{"", "   ", "\n\t"} {
		if _, err := DecodeBase64(in); !errors.Is(err, ErrEmpty) {
			t.Errorf("DecodeBase64(%q) error %v, want ErrEmpty", in, err)
		}
	}
}

func TestDecodeBase64Invalid(t *testing.T) {
	if _, err := DecodeBase64("not base64!"); err == nil {
		t.Fatal("decoding garbage succeeded")
	}
}

func TestDecodeBase64Binary(t *testing.T) {
	// 0xff 0xfe is not UTF-8; the buffer must never see it.
	if _, err := DecodeBase64("//4="); !errors.Is(err, ErrNotText) {
		t.Fatalf("error %v, want ErrNotText", err)
	}
}

func TestBase64RoundTrip(t *testing.T) {
	for _, in := range []string{"hello", "中文", "a\nb", "{}[]:,", strings.Repeat("x", 1000)} {
		got, err := DecodeBase64(EncodeBase64(in))
		if err != nil {
			t.Fatalf("round trip %q: %v", in, err)
		}
		if got != in {
			t.Fatalf("round trip %q → %q", in, got)
		}
	}
}

func TestMD5Forms(t *testing.T) {
	cases := []struct {
		in   string
		mode MD5Mode
		want string
	}{
		{"", MD5Hex32Lower, "d41d8cd98f00b204e9800998ecf8427e"},
		{"", MD5Hex16Lower, "8f00b204e9800998"},
		{"", MD5Hex32Upper, "D41D8CD98F00B204E9800998ECF8427E"},
		{"", MD5Hex16Upper, "8F00B204E9800998"},
		{"abc", MD5Hex32Lower, "900150983cd24fb0d6963f7d28e17f72"},
		{"abc", MD5Hex16Lower, "3cd24fb0d6963f7d"},
		{"abc", MD5Hex32Upper, "900150983CD24FB0D6963F7D28E17F72"},
		{"abc", MD5Hex16Upper, "3CD24FB0D6963F7D"},
		{"中文", MD5Hex32Lower, "a7bac2239fcdcb3a067903d8077c4a07"},
	}
	for _, c := range cases {
		if got := MD5(c.in, c.mode); got != c.want {
			t.Errorf("MD5(%q, %v) = %q, want %q", c.in, c.mode, got, c.want)
		}
	}
}

func TestFormatJSON(t *testing.T) {
	in := `{"b":1,"a":[true,null,"x"],"n":1.2300e2}`
	want := `{
  "b": 1,
  "a": [
    true,
    null,
    "x"
  ],
  "n": 1.2300e2
}`
	got, err := FormatJSON(in)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("FormatJSON =\n%s\nwant\n%s", got, want)
	}
	// Member order and number spelling stay as written.
	if strings.Index(got, `"b"`) > strings.Index(got, `"a"`) {
		t.Fatal("member order changed")
	}
	if !strings.Contains(got, "1.2300e2") {
		t.Fatal("number spelling changed")
	}
}

func TestFormatJSONWhitespaceInput(t *testing.T) {
	got, err := FormatJSON("\n  [1, 2]  \n")
	if err != nil {
		t.Fatal(err)
	}
	if got != "[\n  1,\n  2\n]" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatJSONCRLF(t *testing.T) {
	got, err := FormatJSON("{\r\n\"a\": 1\r\n}")
	if err != nil {
		t.Fatal(err)
	}
	if got != "{\n  \"a\": 1\n}" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatJSONKeepsDuplicateNames(t *testing.T) {
	got, err := FormatJSON(`{"a":1,"a":2}`)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"a\": 1,\n  \"a\": 2\n}"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFormatJSONInvalid(t *testing.T) {
	for _, in := range []string{"", "  ", "{", `{"a":1,}`, `{"a":1} {"b":2}`} {
		if _, err := FormatJSON(in); !errors.Is(err, ErrJSON) {
			t.Errorf("FormatJSON(%q) error %v, want ErrJSON", in, err)
		}
	}
}

func TestFormatJSONDepth(t *testing.T) {
	deep := strings.Repeat("[", MaxJSONDepth+1) + strings.Repeat("]", MaxJSONDepth+1)
	if _, err := FormatJSON(deep); !errors.Is(err, ErrJSONDepth) {
		t.Fatalf("error %v, want ErrJSONDepth", err)
	}
	ok := strings.Repeat("[", MaxJSONDepth) + "1" + strings.Repeat("]", MaxJSONDepth)
	if _, err := FormatJSON(ok); err != nil {
		t.Fatalf("formatting at the depth limit: %v", err)
	}
}

func TestFormatJSONOutputLimit(t *testing.T) {
	// Every value sits at the deepest level, so each line is indented by
	// MaxJSONDepth*2 spaces: the output passes MaxJSONOutput long before
	// the input is close to MaxInput.
	deep := strings.Repeat("[", MaxJSONDepth)
	in := deep + strings.Repeat("0,", 40000) + "0" + strings.Repeat("]", MaxJSONDepth)
	if len(in) > MaxInput {
		t.Fatalf("test input is %d bytes, over MaxInput", len(in))
	}
	if _, err := FormatJSON(in); !errors.Is(err, ErrJSONOutput) {
		t.Fatalf("error %v, want ErrJSONOutput", err)
	}
}

func TestMinifyJSON(t *testing.T) {
	got, err := MinifyJSON(" {\n  \"a\": 1,\n  \"b\": [2, 3]\n} ")
	if err != nil {
		t.Fatal(err)
	}
	if got != `{"a":1,"b":[2,3]}` {
		t.Fatalf("got %q", got)
	}
}

func TestMinifyJSONKeepsStrings(t *testing.T) {
	got, err := MinifyJSON(`{"a": "x y\tz", "b": "q\n"}`)
	if err != nil {
		t.Fatal(err)
	}
	if got != `{"a":"x y\tz","b":"q\n"}` {
		t.Fatalf("got %q", got)
	}
}

func TestMinifyJSONInvalid(t *testing.T) {
	if _, err := MinifyJSON(`{"a":1,}`); !errors.Is(err, ErrJSON) {
		t.Fatalf("error %v, want ErrJSON", err)
	}
}
