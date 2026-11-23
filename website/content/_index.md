---
title: ZPad
hero:
  eyebrow: "v0.3.2 已发布"
  title: "轻量、原生、纯粹的文本编辑器"
  lead: "左右分屏、多标签、Markdown 即时预览。界面全部由 Go 绘制，没有 WebView。"
  note: "macOS 与 Windows · Linux 即将上线"
  primary: "下载"
  secondary: "GitHub"
  shot: split
  shot_dark: split-dark
  alt: "ZPad 主界面：左侧 Markdown，右侧 Go 代码，各占一栏"
features_title: "只做编辑器该做的事"
features_lead: "没有插件、没有云同步、没有 Git、没有 AI、没有 WebView。"
features:
  - icon: native
    title: "轻量低占用"
    text: "原生绘制，启动快，常驻内存远低于 Electron 编辑器。"
  - icon: encoding
    title: "多格式与高亮"
    text: "txt、md、json、go、html、css 双击即开，按扩展名高亮。"
  - icon: split
    title: "左右分屏"
    text: "两栏各开多标签，同一文件共享一份缓冲区。"
  - icon: markdown
    title: "Markdown 预览"
    text: "原生渲染标题、列表、代码块和本地图片。"
  - icon: sun
    title: "亮 / 暗主题"
    text: "跟随系统，也能手动切换，窗口和编辑器一起变。"
  - icon: grid
    title: "开箱即用"
    text: "没有配置向导和插件市场，文件拖进来就能写。"
features_more: "还有编码与行尾（UTF-8 / UTF-16 / GB18030）、查找替换、会话恢复。"
gallery_title: "截图"
gallery:
  - name: split
    alt: "亮色主题下的左右分屏"
    caption: "左右分屏：同一份文件在两栏各自滚动。"
  - name: split-dark
    alt: "暗色主题下的左右分屏"
    caption: "暗色主题，整个窗口一起变。"
  - name: settings
    alt: "设置面板"
    caption: "主题、字体、字号、语言，改完即生效。"
  - name: settings-dark
    alt: "暗色主题下的设置面板"
    caption: "设置面板在暗色下的样子。"
faq_title: "常见问题"
faq:
  - q: "ZPad 收费吗？"
    a: "免费。以 AGPL-3.0 开源，源码和每个版本都在 GitHub 上。"
  - q: "会上传我的文件吗？"
    a: "不会。应用不联网也能正常使用，唯一的网络请求是向 GitHub 检查更新。"
  - q: "有统计或追踪吗？"
    a: "没有。代码完全开源，不含任何追踪和匿名数据上报；官网只使用 Cloudflare 的无 cookie 访问统计。"
  - q: "为什么 macOS 或 Windows 会拦截？"
    a: "安装包没有 Apple / Microsoft 的开发者签名，系统会按默认策略先拦一下：macOS 右键打开，Windows 选择「仍要运行」。介意的话可以自己核对源码，或者从源码构建。"
  - q: "自动更新安全吗？"
    a: "更新包用 Ed25519 签名校验，签名不符不会安装。应用需要装在可写目录才能自更新。"
  - q: "支持哪些系统？"
    a: "macOS 12 及以上、Windows 10 / 11（64 位）。Linux 正在适配。"
  - q: "为什么不支持插件、Git、AI？"
    a: "定位是一个纯粹的文本编辑器：启动够快、行为可预期。需要这些能力时，用更重的 IDE 更合适。"
changelog_title: "最近更新"
changelog_lead: "每次发布都会写在 GitHub Releases 里。"
changelog_link: "全部版本"
---
