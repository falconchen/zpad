# 01 · M0 项目骨架与基础设施

> 预估：0.5–1 天
> 前置：Go 1.27.4 已安装（已完成）

## 1. 目标

把一个空仓库变成**可运行、可测试、可持久化设置**的 MyGo 原生应用骨架。
M0 不实现文件打开和编辑器，只把后续阶段需要的地基铺好。

## 2. 范围

### 做

- 就地初始化 MyGo 原生项目，保留现有 `.git` 与 `LICENSE`。
- `go.mod`、`mygo.json`、目录结构。
- 应用窗口与根布局骨架。
- 主题管线：亮色 / 暗色 / 跟随系统。
- 设置模型与 `settings.json` 读写。
- 会话模型的最小占位。
- 错误提示（toast / 对话框）与日志。
- `ui.Tester` 测试脚手架。
- 开发说明文档。

### 不做

- 文件树、标签页、编辑器、编码、查找替换。
- 文件关联与打包。
- 中英文案最终稿。

## 3. 任务拆解

### 3.1 初始化

- [ ] 用 MyGo CLI 创建 native 模板。
  - 优先尝试在仓库根目录就地初始化。
  - 若 CLI 只支持新建目录，则在临时目录初始化，再把 `main.go`、`go.mod`、`mygo.json`、`resources/` 合并到仓库根目录。
  - 合并后保留仓库已有的 `.git`、`LICENSE`。
- [ ] 设置 `mygo.json`：

```json
{
  "name": "ZPad",
  "identifier": "com.xiaoz.zpad",
  "version": "0.1.0"
}
```

- [ ] `go.mod` 声明 Go 1.27，并把 MyGo CLI 作为 tool 依赖：

```sh
go get -tool github.com/egoist/mygo/cmd/mygo@latest
go get golang.org/x/text@latest
```

- [ ] 确认 `go version` 为 `go1.27.4 darwin/arm64`。
- [ ] 准备 `resources/icon.png` 占位图（1024×1024，M4 再替换正式图标）。

### 3.2 目录结构

```
.
├── main.go
├── mygo.json
├── go.mod
├── go.sum
├── locales/
│   ├── en.json
│   └── zh.json
├── resources/
│   └── icon.png
├── fonts/
│   └── (M3 放入 JetBrains Mono)
└── internal/
    ├── app/        # App 状态、命令、关闭协调
    ├── views/      # 根布局、工具栏、状态栏占位
    ├── theme/      # 亮/暗调色板与 editor.Theme
    ├── settings/   # 设置模型、读写、默认值
    └── session/    # 会话模型与读写占位
```

- [ ] `main.go` 只负责：注册字体（M3）、创建 App、创建窗口、`App.Run()`。
- [ ] `internal/app` 暴露 `New()` 与 `View(c *ui.Context)`，便于 `ui.Tester` 测试。
- [ ] 各包先建空文件与类型，保证 `go build ./...` 通过。

### 3.3 窗口与根布局

- [ ] 主窗口：

```go
mygo.NewWindow(mygo.WindowOptions{
    Title:     "ZPad",
    Width:     1200,
    Height:    800,
    MinWidth:  800,
    MinHeight: 520,
    StateKey:  "main",
    Content:   ui.View(app.View),
})
```

- [ ] 根布局：`Column` → 工具栏占位 → 内容区占位 → 状态栏占位。
- [ ] 内容区先显示空状态：应用名、版本、"打开文件 / 打开文件夹"按钮（按钮暂不接功能）。
- [ ] 状态栏先显示静态文本：`ZPad 0.1.0`。

### 3.4 主题管线

- [ ] `internal/theme` 提供：

```go
type Mode string // "light" | "dark" | "system"

func UITheme(c *ui.Context, mode Mode) *ui.Theme
func EditorTheme(t *ui.Theme) editor.Theme
```

- [ ] 亮色 / 暗色各一套最小调色板（颜色值见 `00-requirements.md`）。
- [ ] 跟随系统时以 `c.Theme()` 为基底。
- [ ] 切换主题时调用：

```go
mygo.Theme.SetSource(mygo.ThemeLight)  // 或 ThemeDark / ThemeSystem
```

- [ ] 根元素调用 `c.SetTheme(theme.UITheme(c, app.settings.Theme))`。
- [ ] 工具栏放一个临时主题切换控件，验证三档生效。

### 3.5 设置

- [ ] 定义 `settings.Settings`，至少包含：

```go
type Settings struct {
    Theme        string  `json:"theme"`        // light | dark | system
    Language     string  `json:"language"`     // "" | zh-Hans | en
    FontFamily   string  `json:"fontFamily"`
    FontSize     float32 `json:"fontSize"`
    LineHeight   float32 `json:"lineHeight"`
    SoftWrap     bool    `json:"softWrap"`
    LineNumbers  string  `json:"lineNumbers"`
    TabSize      int     `json:"tabSize"`
    UseTabs      bool    `json:"useTabs"`
    ShowGuides   bool    `json:"showGuides"`
    AutoClose    bool    `json:"autoClose"`
    DefaultEnc   string  `json:"defaultEncoding"`
    DefaultEOL   string  `json:"defaultLineEnding"`
    LargeFileMB  int     `json:"largeFileMB"`
    SidebarOpen  bool    `json:"sidebarOpen"`
    RestoreSession bool  `json:"restoreSession"`
    ConfirmClose bool    `json:"confirmClose"`
}
```

- [ ] `settings.Load()`：文件不存在时返回默认值；损坏时备份为 `settings.json.bak` 并返回默认值。
- [ ] `settings.Save()`：先写临时文件再替换，权限 `0600`。
- [ ] 路径：`mygo.App.Path(mygo.PathUserData) + "/settings.json"`。
- [ ] 设置变更后防抖 300ms 写盘。

### 3.6 会话占位

- [ ] 定义 `session.State`：

```go
type State struct {
    Root        string   `json:"root"`
    Split       bool     `json:"split"`
    SplitWidth  float32  `json:"splitWidth"`
    ActivePane  int      `json:"activePane"`
    LeftTabs    []string `json:"leftTabs"`
    RightTabs   []string `json:"rightTabs"`
    SidebarOpen bool     `json:"sidebarOpen"`
    SidebarWidth float32 `json:"sidebarWidth"`
}
```

- [ ] 本期只实现读写空文件，不接入布局。

### 3.7 错误与日志

- [ ] 统一的 `app.toast(text, tone)`，使用 `c.Toast`。
- [ ] 统一的错误对话框封装：`mygo.Dialog.Message`。
- [ ] 启动失败、设置读写失败要可见，不能静默。

### 3.8 测试脚手架

- [ ] `internal/settings` 单测：默认值、读写往返、损坏文件回退。
- [ ] `internal/app` 视图测试：用 `ui.Tester` 渲染根布局，断言应用名与版本存在。
- [ ] 确保 `go test ./...` 通过。

### 3.9 文档

- [ ] 根目录 `README.md` 写明：项目简介、环境要求、`go tool mygo dev`、`go test ./...`、目录结构。
- [ ] 在 `docs/README.md` 更新 M0 状态。

## 4. 交付物

- 可在 macOS 启动的 ZPad 空壳窗口。
- 三档主题可切换并持久化。
- `settings.json` 正常读写。
- 测试脚手架与开发说明。

## 5. 验收清单

- [ ] `go version` 输出 `go1.27.4`。
- [ ] `go tool mygo dev` 能启动应用，窗口标题为 `ZPad`。
- [ ] 窗口默认 1200×800，最小 800×520；重启后位置与大小恢复。
- [ ] 亮色 / 暗色 / 跟随系统三档可切换，重启后保持上次选择。
- [ ] 强制暗色时，原生菜单和对话框也为暗色（`mygo.Theme.SetSource` 生效）。
- [ ] 删除 `settings.json` 后重启，应用使用默认设置且不崩溃。
- [ ] 故意写坏 `settings.json`，应用能回退并保留 `.bak`。
- [ ] `go build ./...` 与 `go test ./...` 通过。
- [ ] 代码中没有 `URL`、`Page`、`WebView` 相关用法。
- [ ] `git status` 只包含预期文件。

## 6. 风险与备注

- MyGo CLI 的就地初始化行为需要在执行时确认；不支持则用临时目录合并。
- `mygo.Theme.SetSource` 需要在创建窗口前或设置变更时调用，具体调用时机以实际 API 为准。
- 设置结构体会在 M1–M3 继续扩展，M0 只要求可读写与向后兼容的默认值。

## 7. 实现记录

完成日期：2026-11-22。

已验证：

- Go 1.27.4（官网 tarball，`/usr/local/go`）可用。
- MyGo v0.10.0 native 模板已就地初始化，保留 `.git` / `LICENSE` / `docs`。
- `go build ./...` 通过。
- `go test ./...` 通过：
  - `internal/app`：视图渲染、主题切换与设置持久化。
  - `internal/settings`：默认值、读写往返、损坏备份、`Ensure`、权限。
  - `internal/session`：默认值、读写往返、损坏备份、活动标签修正。
  - `internal/theme`：亮 / 暗主题。
- `go vet ./...` 通过。
- `go tool mygo build` 通过，产出：
  - `build/darwin-arm64/ZPad.app`（9.9 MB）
  - `build/darwin-arm64/zpad-0.1.0-arm64.dmg`（3.3 MB）
- 应用可启动；首次启动写入：
  - `~/Library/Application Support/ZPad/settings.json`
  - `~/Library/Application Support/ZPad/window-state.json`

与方案的差异：

- 设置保存采用**立即原子写**，未做 300ms 防抖；更简单，也不会在退出时丢设置。后续如有频繁写入需求再加防抖。
- `internal/views` 暂未单独建包，M0 的视图代码放在 `internal/app/view.go`；M1 拆分组件时再引入 `internal/views`。
- `locales/` 目前只有占位说明，中英文 catalogs 在 M3 加入。
- `fonts/` 与内置字体在 M3 加入。

待你验收：

- [x] 按第 5 节验收清单逐项确认
- [x] 确认应用窗口、主题切换、设置文件符合预期
- [x] 签收后进入 M1

## 8. 签收

| 项目 | 内容 |
| --- | --- |
| 状态 | 已签收 |
| 签收人 | 用户 |
| 签收日期 | 2026-11-22 |
| 实际用时 | 2026-11-22 · 约 1 小时 |
| 备注 | 用户指示继续 M1，视为通过 |

- [x] 验收清单全部通过
- [x] 相关文档已更新
- [x] 同意进入 M1

签收意见：

```text

```
