---
title: "Download"
heading: "Download ZPad"
layout: download
lead: "Available for macOS and Windows. Linux is on the way. Install once, it updates itself."
macos_title: "macOS 12 or later"
macos_lead: "Pick the build for your chip — the wrong one will not open."
windows_title: "Windows 10 / 11 (64-bit)"
windows_lead: "The installer registers ZPad for text files; the portable zip runs from anywhere."
linux_title: "Linux"
linux_lead: "Being ported. It will show up here once it is stable."
notes_title: "Install notes"
notes:
  - text: "On macOS, Gatekeeper blocks the first launch: right-click the app and choose Open, or allow it under System Settings → Privacy & Security."
  - text: "If it still says the app is damaged or cannot be verified, the installer is not signed by Apple, so remove the quarantine flag by hand. Installing in /Applications usually needs sudo."
    code: "sudo xattr -dr com.apple.quarantine /Applications/ZPad.app"
  - text: "On Windows, SmartScreen warns about an unknown publisher: More info → Run anyway."
  - text: "The updater needs the app in a writable folder (such as ~/Applications). Running it straight from the read-only DMG will not update."
license_note: "ZPad is open source under AGPL-3.0; the source and every release live on GitHub."
---
