# 官网（zpad.xphub.dev）

ZPad 的产品官网：首屏、功能、截图、常见问题、下载。源码在 `website/`，用 Hugo 生成静态文件，部署在 Cloudflare Pages。

- 线上地址：<https://zpad.xphub.dev>
- 默认语言中文（根路径），英文在 `/en/`
- 下载按钮读发布数据，不需要改代码
- 没有 npm、没有前端框架、没有构建服务端

## 范围

首页五段：首屏（截图 + 下载 / GitHub 按钮）、六个功能点、四张截图（两列，点击放大）、常见问题、最近更新。导航是 功能 / 下载 / 更新 / 反馈 / GitHub，反馈直接进 GitHub 新建 issue。下载页按平台列安装包，Linux 显示「即将上线」。不做的：博客、文档站、评论区、账号体系。

`docs/` 下的开发文档不对外，官网只链到 GitHub。

## 技术选型

| 项目 | 选择 | 原因 |
| --- | --- | --- |
| 生成器 | Hugo **0.170.1** | 单二进制、无 npm，CF Pages 有官方预设；多语言内建 |
| 样式 | 手写 CSS（`assets/css/site.css`） | 页面不多，用不上 UI 框架；色板直接沿用 `internal/theme` |
| 字体 | 系统字体 | 首屏不下载字体文件 |
| 图标 | 内联 SVG（`layouts/partials/icon.html`）+ 应用图标 PNG | 不引外部图标库；favicon 和页头 logo 用应用自己的 icon |
| 主题 | 亮 / 暗，跟随系统，可手动切换 | 和 App 一致；首屏截图跟着主题换 |
| 截图 | `go test` 里的 headless 渲染 | 不需要屏幕录制权限，中英 / 亮暗各一套，可复现 |

配色取自 `internal/theme`：亮色 `#4F46E5` 主色、`#F7F8FA` 表面色；暗色 `#0F1115` 背景、`#7C8CFF` 主色。

## 目录

```
website/
├── hugo.toml                  # 站点、语言、输出设置
├── content/
│   ├── _index.md|.en.md       # 首页文案：功能点、截图说明、FAQ
│   └── download.md|.en.md     # 下载页文案与安装说明
├── i18n/{zh,en}.toml          # 模板里复用的 UI 字符串
├── layouts/
│   ├── baseof.html            # 含截图灯箱 <dialog>
│   ├── index.html             # 首页
│   ├── _default/download.html # 下载页（layout: download）
│   ├── robots.txt
│   └── partials/{head,header,footer,icon,shot}.html
├── assets/
│   ├── css/ js/
│   ├── icon.png               # resources/icon.png 裁掉留白的副本
│   └── screenshots/*.png      # 截图原图，Hugo 构建时转 WebP
├── static/_redirects          # 由脚本生成，勿手改
├── data/release.json          # 由脚本生成，勿手改
└── scripts/release/main.go    # 生成上面两个文件
```

## 多语言

- `defaultContentLanguage = "zh"`、`defaultContentLanguageInSubdir = false`：中文在 `/`，英文在 `/en/`
- 页面翻译按文件名：`content/_index.md`（中）+ `content/_index.en.md`（英）
- 模板里的固定文案走 `i18n`，页面文案走各自语言 front matter
- `head.html` 输出 `hreflang`（含 `x-default` 指中文），语言切换链接指向当前页面的对应译文
- 截图按语言分两套：中文 `split.png`，英文 `split-en.png`，由 front matter 里的名字决定

## 图标与 favicon

`resources/icon.png` 是应用图标（1024×1024，圆角外透明）。构建期用 Hugo 的图片处理生成 32 / 192 / 180（apple-touch-icon）三个 PNG，页头 logo 用同一张图的 28px 版本，所以官网和 App 的图标永远一致。

`website/assets/icon.png` 是从 `resources/icon.png` 裁掉约 10% 透明留白后的副本（824×824）。**换应用图标时要同步这个文件**，否则官网还是旧图标：

```sh
sips -c 824 824 resources/icon.png --out website/assets/icon.png
```

## 截图

截图不是手截的，由 `internal/app/site_shots_test.go` 在内存里渲染，因此不需要屏幕录制权限，也不会混进真实路径：

```sh
ZPAD_SHOTS=/tmp/shots go test ./internal/app -run TestSiteScreenshots
cp /tmp/shots/*.png website/assets/screenshots/     # 先按需缩小
```

渲染出 2 个场景 × 亮暗 × 中英 = 8 张，1600×900 左右：

| 场景 | 内容 |
| --- | --- |
| `split` | 左右分屏：左边 Markdown 源码，右边 Go 代码 |
| `settings` | 设置面板 |

Markdown 预览的截图暂时不放：预览的呈现方式还要调整，官网不展示一个会变的东西。功能卡里也只写「原生渲染标题、列表、代码块和本地图片」，不承诺具体布局。

场景由测试里的 demo 工作区（`zpad-demo`，5 个文件）拼出来，改文案或改功能后重跑一次即可。原图按 `sips -Z 1600` 缩到 1600 宽再放进 `assets/`，Hugo 构建时转成 WebP（`shot.html`），整站产物约 550 KB。

README 里的截图用同一套：`docs/screenshots/editor-light.png`、`editor-dark.png`、`settings-dark.png` 分别是 `split`、`split-dark`、`settings` 的副本，重截后一起覆盖。

首屏用 `split` + `split-dark`，跟随站点主题切换；截图区固定展示 4 张（亮暗各两张），点击用原生 `<dialog>` 放大：控制条在图片下方居中（上一张 / 张数 / 下一张 / 关闭），支持 Esc、点击遮罩、左右方向键，脚本不跑时链接直接打开原图。首屏大图不参与灯箱，点它不会弹出。

## 下载数据

首屏胶囊（`v0.4.0 已发布`）与首页「最近更新」都从 `data/release.json` 生成，`i18n` 里只有 `is_out` 这个词。不要在 `content/*.md` 的 front matter 里写版本号：0.3.2 那版就是手写在 `hero.eyebrow` 里，0.4.0 发布后官网还挂着旧版本号；`release_test.go` 的 `TestWebsiteHasNoHardcodedVersion` 现在会拦住这种写法。

`data/release.json` 的结构：

```json
{
  "tag": "v0.3.2",
  "version": "0.3.2",
  "published_at": "2026-11-23T05:13:44Z",
  "date": "2026-11-23",
  "url": "https://github.com/helloxz/zpad/releases/tag/v0.3.2",
  "notes": "* 优化：根据操作系统动态调整快捷键提示",
  "assets": {
    "macos_arm64":      { "name": "…", "url": "…", "size": "5.8 MB", "bytes": 6040064 },
    "macos_amd64":      { … },
    "windows_setup":    { … },
    "windows_portable": { … },
    "checksums":        { … }
  }
}
```

`.github/workflows/website.yml` 在 `release: published` 时调用 GitHub API 拿到 release，跑 `website/scripts/release/main.go` 重写 `data/release.json` 和 `static/_redirects`，然后提交。CF Pages 的 Git 集成看到这次提交就重新构建，所以：

- 发版流程不变：打 tag → 等 CI 把 draft 的产物传完 → 发布为 latest
- 官网 1 分钟内自动跟上新版本
- 构建期不访问 GitHub API，不需要在 CF 里放 token
- 手动补数据：Actions 里跑一次 `Website`（`workflow_dispatch`，默认取 latest）

**发布时机很重要**：作业由 `release: published` 触发，读到的 release 必须已经有文件。draft 在 macOS / Windows 构建期间陆续填充，几分钟才齐；先点 Publish、构建还在跑，作业会读到 `assets: []` 并以 `release vX.Y.Z has no files` 失败（0.3.3、0.4.0 都这样失败过一次）。补救办法是等 Release workflow 变绿后在 Actions 里跑一次 `Website`。

固定短链（生成到 `_redirects`，可写进 README 或别人的博客）：

```
/download/mac-arm64    → 最新 Apple Silicon dmg
/download/mac-intel    → 最新 Intel dmg
/download/windows      → 最新安装包
/download/windows-portable
```

`.delta` 和 `update-*.json` 是更新器的产物，脚本按文件名过滤，不会出现在下载页。发布但没勾选 latest、或者发成 pre-release 时，脚本会跳过（`draft`/`prerelease` 直接退出），避免官网指向更新器不提供的版本。

## 部署（Cloudflare Pages）

1. Workers & Pages → Create → Pages → Connect to Git → 选 `helloxz/zpad`
2. 构建配置：

   | 字段 | 值 |
   | --- | --- |
   | Production branch | `main` |
   | Framework preset | Hugo |
   | Root directory | `website` |
   | Build command | `hugo --minify` |
   | Build output directory | `public` |

3. 环境变量 `HUGO_VERSION=0.170.1`。CF 的默认 Hugo 版本会漂，锁住才能和本地一致。
4. Custom domains → 填 `zpad.xphub.dev`，DNS 加一条 CNAME 指向 `<project>.pages.dev`。
5. 部署后确认 `/download/mac-arm64` 会 302 到当前版本的 dmg。

`website/public` 不进 Git，CF 自己构建。

## Web Analytics

Pages 项目 → Metrics → Web Analytics → 开启，拿到 token 后填到 `website/hugo.toml`：

```toml
[params]
  cfAnalyticsToken = "<token>"
```

`head.html` 只在 token 非空时渲染 beacon，没填就没有任何统计脚本。无 cookie，不需要同意弹窗。FAQ 里「有统计或追踪吗」的回答要与此保持一致：应用零统计，官网只有这份无 cookie 的聚合访问量。

## 本地开发

```sh
brew install hugo                  # 0.170.1
cd website && hugo server          # http://localhost:1313/
hugo --minify                      # 产出 website/public
```

只在 `website/` 下改，`go test ./...` 不会碰它。

## 验收清单

- [ ] 首页首屏显示版本号、截图、下载 / GitHub 按钮，logo 是应用图标
- [ ] 语言切换在首页和下载页都能回到对应译文
- [ ] 亮 / 暗切换后首屏截图跟着换，刷新后保持
- [ ] 截图区四张、两列排布，点击后放大，按钮在图片下方居中，Esc / 遮罩 / 方向键可用
- [ ] 常见问题折叠展开正常，内容与隐私表述一致
- [ ] 手机（390px）下导航、卡片、按钮、截图不溢出，截图单列
- [ ] 导航的「反馈」打开 GitHub 新建 issue 页面
- [ ] 下载页四个按钮指向当前版本，Linux 显示「即将上线」，安装说明含 sudo 命令
- [ ] `/download/mac-arm64` 302 到当前版本 dmg
- [ ] 新版本发布后 5 分钟内，官网版本号与下载链接更新
- [ ] 无 JS 时页面仍可读、按钮仍可用（首屏截图固定暗色，截图链接直接开原图）

## 实现记录

- 2026-11-23：Hugo 站点落地。首页、下载页、双语、亮暗主题、发布数据自动化、CF Pages 接入说明。
- 2026-11-23：首屏与截图区改用 headless 渲染的新截图（中英各一套）；favicon 与页头 logo 换成 `resources/icon.png`；新增常见问题；下载页改名「安装说明」并补 `xattr` 命令；截图支持点击放大。
- 2026-11-23：标题定为「轻量、原生、纯粹的文本编辑器」；撤掉 Markdown 预览截图，图库缩为四张两列（预览呈现方式待调整）；灯箱控制条移到图片下方居中并加张数；导航加「反馈」进 GitHub issues；`word-break: keep-all` 避免手机端把「纯粹」拆行。
- 未接入统计：`cfAnalyticsToken` 留空，拿到 token 填入即生效。
