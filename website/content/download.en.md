---
title: "Download"
heading: "Download ZPad"
layout: download
lead: "Available for macOS, Windows and Linux. Pick the right package; install and update notes are below."
macos_title: "macOS 12 or later"
macos_lead: "Pick the build for your chip — the wrong one will not open."
windows_title: "Windows 10 / 11 (64-bit)"
windows_lead: "The installer registers ZPad for text files; the portable zip runs from anywhere."
linux_title: "Linux"
linux_lead: "x86_64 deb and rpm: deb for Debian / Ubuntu, rpm for Fedora / RHEL."
notes_title: "Install notes"
notes:
  - text: "On macOS, Gatekeeper blocks the first launch: right-click the app and choose Open, or allow it under System Settings → Privacy & Security."
  - text: "If it still says the app is damaged or cannot be verified, the installer is not signed by Apple, so remove the quarantine flag by hand. Installing in /Applications usually needs sudo."
    code: "sudo xattr -dr com.apple.quarantine /Applications/ZPad.app"
  - text: "On Windows, SmartScreen warns about an unknown publisher: More info → Run anyway."
  - text: "On Linux the deb and rpm install into /opt and update through the package manager, not the app itself; other distributions can use the tar.gz from the release."
  - text: "To get in-app updates, install with install.sh into ~/.local:"
    code: "curl -fsSL https://github.com/helloxz/zpad/releases/latest/download/install.sh | sh"
  - text: "The updater needs the app in a writable folder (such as ~/Applications on macOS); an app run from the read-only DMG, or installed machine-wide in Program Files, will not update."
license_note: "ZPad is open source under AGPL-3.0; the source and every release live on GitHub."
---
