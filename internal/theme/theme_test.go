package theme

import "testing"

func TestFor(t *testing.T) {
	light, dark := For(false), For(true)
	if light.Dark {
		t.Fatal("light theme is dark")
	}
	if !dark.Dark {
		t.Fatal("dark theme is not dark")
	}
	if light.Background == dark.Background {
		t.Fatal("light and dark share a background")
	}
	if light.Radius != 8 || dark.Radius != 8 {
		t.Fatalf("radii %v, %v; want 8", light.Radius, dark.Radius)
	}
	ct := EditorTheme(light)
	if ct.Background != light.Background || ct.Text != light.Text {
		t.Fatal("editor theme does not follow the UI theme")
	}
}
