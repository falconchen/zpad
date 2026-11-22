# AGENTS.md

ZPad 是一个基于 **MyGo 原生 UI** 的轻量级文本 / 代码编辑器。本文件给后续接手本项目的 agent 提供核心上下文，详细需求与里程碑见 `docs/`。

## 项目概况

- 仓库：`https://github.com/helloxz/zpad.git`，分支 `main`
- Module：`github.com/helloxz/zpad`
- 应用名：**ZPad**
- Bundle ID：**`com.xiaoz.zpad`**
- 版本：`0.1.0`
- 许可证：**AGPL-3.0**（对外分发前确认源码义务）
- 平台顺序：**macOS 先跑通 → Windows（后续）→ Linux（最后）**
- 当前状态：**M3 已完成，等待签收**；M4 未开始

## 技术栈

- Go **1.27.4**，官网 tarball 安装，`/usr/local/go`，`GOROOT=/usr/local/go`
- MyGo **v0.10.0**：`github.com/egoist/mygo`
- UI：`package ui` 原生绘制，**不使用 WebView**
- 编辑器：`github.com/egoist/mygo/ui/editor`
- 编码：`golang.org/x/text`（纯 Go，无 cgo）
- 字体：内置 JetBrains Mono（OFL），`ui.RegisterFont` + `ui.FontFamilies`
- 国际化：MyGo catalogs，`locales/en.json` + `locales/zh-Hans.json`，生成 `locales/catalog_gen.go`
- CLI：项目内固定为 tool，使用 `go tool mygo`

常用命令：

```sh
go tool mygo dev     # 开发版，改 .go 自动重启
go test ./...        # 单元测试 + ui.Tester 视图测试
go vet ./...
go tool mygo generate  # 改了 locales/*.json 后重新生成 catalog_gen.go
go tool mygo build   # 打包 macOS .app / .dmg
```

产物：`build/darwin-arm64/ZPad.app`、`zpad-0.1.0-arm64.dmg`。

## 产品需求（一期）

- 单窗口；左侧树形目录（默认关闭，打开文件夹时展开）；内容区左右双栏；底部状态栏。
- 每栏多标签；同一文件跨栏共享 `Buffer`，光标 / 滚动独立。
- 编辑 HTML、CSS、JS、JSON、Markdown、纯文本。
- 查找替换（内建 FindBar）、大小写转换（大写 / 小写 / 首字母大写 / 反转，多选区）。
- 编码：UTF-8、UTF-8 BOM、UTF-16 LE/BE、GB18030；自动检测 +「重新以编码打开 / 以编码保存」。
- 换行符：LF / CRLF / CR，保存保留原行尾。
- 主题：亮色 / 暗色 / 跟随系统；中英双语（M3）；内置 JetBrains Mono + 系统等宽字体（M3）。
- 手动保存 + 未保存标记；关闭标签提示；关闭窗口 / 退出汇总提示「保存全部 / 不保存 / 取消」。
- 会话恢复：工作区、分屏、标签、活动标签、滚动、侧栏；窗口位置由 MyGo `StateKey: "main"` 管理。
- 文件关联、拖拽打开；文件关联在 M4 完成。

非目标：LSP、调试器、Git 集成、插件系统、云同步、HTML 实时预览、移动端。Markdown 预览与全局搜索放二期。

## 代码结构

```
main.go                    入口：设置、主题源、窗口、OnOpenFile、OnBeforeQuit
internal/app/
  app.go                   状态、pane/tab、分屏、主题、窗口标题
  document.go              文档加载 / 保存、编码、二进制与大小拦截
  actions.go               打开 / 保存 / 关闭标签 / 编码 / 换行 / 大小写
  close.go                 关闭 / 退出协调、脏文档、保存全部
  session.go               会话保存与恢复
  view.go                  工具栏、侧栏、双栏、标签、状态栏、查找栏样式
  settings_ui.go           设置面板、字体列表、语言切换
  menu.go                  应用菜单
  case.go                  多选区大小写转换
fonts.go                   嵌入并注册 JetBrains Mono
fonts/                     JetBrains Mono TTF + OFL
locales/                   中英文 catalogs + 生成的 catalog_gen.go
internal/encoding/         编码检测 / 解码 / 编码 / 往返校验
internal/files/            工作区目录树
internal/caseconv/         大小写转换
internal/settings/         设置模型、JSON 持久化
internal/session/          会话模型、原子读写
internal/theme/            UI 调色板 + 编辑器 token 主题
internal/atomicfile/       原子写文件
docs/                      需求、里程碑、实现记录与签收
```

关键实现约定：

- `editor.Buffer` **只接受 UTF-8**；所有文件编码转换必须走 `internal/encoding`。
- 同一文件在左右栏时复用同一个 `*Document`；每个标签各自持有 `editor.State`。
- 所有 UI / App 状态修改在主线程；文件读写与原生对话框在 goroutine，完成后用 `win.Update` 回到主线程。
- 脏文档按 `*Document` 去重；关闭协调器用 `closePromptOpen` / `allowClose` / `allowQuit` 防止重复弹窗。
- 保存使用原子写；保留原文件权限、编码、BOM 和行尾。
- 会话只保存有路径的文件；未命名文件不持久化。
- 强制外观用 `mygo.Theme.SetSource`，UI 主题用 `c.SetTheme` / `ui.Themed`。
- 不引入 WebView。

## 当前进度

- **M0 项目骨架**：已签收。
- **M1 单栏编辑闭环**：已签收。打开文件 / 文件夹、侧栏、标签、编辑器、保存、编码、换行、大小写、状态栏、菜单、拖拽。
- **M2 左右双栏与会话恢复**：已签收。分屏、跨栏共享 Buffer、会话恢复、关闭 / 退出汇总提示。
- **M3 视觉、设置、字体与中英双语**：已实现，待签收。完整主题、设置面板、JetBrains Mono、中英切换、查找栏 / 跳行栏样式。
- **M4**：大文件只读、编码边界、文件关联、macOS 打包发布（未开始）。

测试覆盖：编码、行尾、大小写、文件树、文档读写、分屏合并、跨栏共享、脏文档去重、会话往返、缺失文件跳过、滚动恢复、无修改退出保存会话、设置面板渲染、字体过滤、中英文案、编辑器 token 颜色。

## 已知限制 / 后续

- 超过设置阈值（默认 20 MB）的文件当前直接拒绝；M4 改为只读打开。
- 标签中键关闭、拖拽排序未实现。
- 设置修改立即写盘，未做防抖。
- 全局搜索、Markdown 预览、外部修改监听、热退出在二期。
- 关闭 / 退出对话框仍需人工点选验证。
- 分隔线宽度在切换分屏 / 关闭时保存，不在拖拽每一帧写盘。
- 编码底层的错误详情可能仍为英文技术信息；主提示已本地化。

## 工作流程

- 先读 `docs/README.md` 与对应里程碑文档；实现完成后在文档中补「实现记录」，状态改为「待签收」。
- 未获得用户明确签收前，不自行把里程碑标记为「已签收」。
- 修改后至少运行 `go test ./...`、`go vet ./...`、`go build ./...`；涉及打包时运行 `go tool mygo build`。
- UI 变更优先用 `ui.Tester` 做视图测试，必要时生成截图人工检查。
- 用户可见文案必须加到 `locales/en.json` 和 `locales/zh-Hans.json`，然后运行 `go tool mygo generate`，通过 `locales.X()` 使用；不要硬编码。
- 新增字体必须确认 OFL / 商业许可，并把许可文件放进 `fonts/` 与 `resources/`。
- 主题颜色优先走 `internal/theme`，不要在 view 里散落硬编码色值。
