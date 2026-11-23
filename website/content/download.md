---
title: "下载"
heading: "下载 ZPad"
layout: download
lead: "macOS、Windows 与 Linux 都能用。选对包，安装和更新方式见下方说明。"
macos_title: "macOS 12 及以上"
windows_title: "Windows 10 / 11（64 位）"
linux_title: "Linux"
notes_title: "安装说明"
notes:
  - text: "macOS 首次打开会被 Gatekeeper 拦下：右键应用 → 打开，或到「系统设置 → 隐私与安全性」里选「仍要打开」。"
  - text: "仍提示「已损坏」或「无法验证开发者」：安装包没有 Apple 开发者签名，需要手动去掉隔离属性。装在 /Applications 时一般要加 sudo。"
    code: "sudo xattr -dr com.apple.quarantine /Applications/ZPad.app"
  - text: "Windows 会提示「未知发布者」：点「更多信息」→「仍要运行」。"
  - text: "Linux 的 deb / rpm 装到 /opt，由包管理器更新，不参与应用内自更新；其他发行版可以用 Releases 里的 tar.gz 压缩包。"
  - text: "想用应用内自更新，就用 install.sh 装到 ~/.local："
    code: "curl -fsSL https://github.com/helloxz/zpad/releases/latest/download/install.sh | sh"
  - text: "自动更新要求应用装在可写目录（例如 macOS 的 ~/Applications）；从只读 DMG 或 Windows 的 Program Files 机器级安装收不到更新。"
license_note: "ZPad 以 AGPL-3.0 开源，源码和每个版本都在 GitHub 上。"
---
