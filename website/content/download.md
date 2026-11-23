---
title: "下载"
heading: "下载 ZPad"
layout: download
lead: "macOS 与 Windows 都能用，Linux 正在适配。装一次，之后自动更新。"
macos_title: "macOS 12 及以上"
macos_lead: "按芯片选择，选错架构会打不开。"
windows_title: "Windows 10 / 11（64 位）"
windows_lead: "安装版会登记为文本文件的打开方式，便携版解压即用。"
linux_title: "Linux"
linux_lead: "正在适配，稳定后出现在这里。"
notes_title: "安装说明"
notes:
  - text: "macOS 首次打开会被 Gatekeeper 拦下：右键应用 → 打开，或到「系统设置 → 隐私与安全性」里选「仍要打开」。"
  - text: "仍提示「已损坏」或「无法验证开发者」：安装包没有 Apple 开发者签名，需要手动去掉隔离属性。装在 /Applications 时一般要加 sudo。"
    code: "sudo xattr -dr com.apple.quarantine /Applications/ZPad.app"
  - text: "Windows 会提示「未知发布者」：点「更多信息」→「仍要运行」。"
  - text: "自动更新要求应用装在可写目录（例如 ~/Applications，或用户目录下的 Applications）；直接从只读 DMG 运行收不到更新。"
license_note: "ZPad 以 AGPL-3.0 开源，源码和每个版本都在 GitHub 上。"
---
