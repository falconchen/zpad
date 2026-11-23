package app

import (
	"runtime"
	"strings"
)

// shortcutText writes a shortcut as the platform's menus write it, as
// shortcut tooltips show it: "CmdOrCtrl+Shift+O" is "⌘⇧O" on macOS and
// "Ctrl+Shift+O" elsewhere. spec uses MyGo's accelerator names, as in
// mygo.json and menu.go.
func shortcutText(spec string) string {
	return shortcutTextFor(runtime.GOOS, spec)
}

// shortcutTextFor writes spec for goos. It is separate so a test on any
// platform can check every platform's text.
func shortcutTextFor(goos, spec string) string {
	parts := strings.Split(spec, "+")
	if len(parts) == 0 {
		return ""
	}
	mac := goos == "darwin"
	var b strings.Builder
	for _, mod := range parts[:len(parts)-1] {
		switch mod {
		case "CmdOrCtrl":
			if mac {
				b.WriteString("⌘")
			} else {
				b.WriteString("Ctrl+")
			}
		case "Alt":
			if mac {
				b.WriteString("⌥")
			} else {
				b.WriteString("Alt+")
			}
		case "Shift":
			if mac {
				b.WriteString("⇧")
			} else {
				b.WriteString("Shift+")
			}
		default:
			b.WriteString(mod)
			b.WriteByte('+')
		}
	}
	b.WriteString(parts[len(parts)-1])
	return b.String()
}
