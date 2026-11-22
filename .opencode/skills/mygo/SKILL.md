---
name: MyGo 原生 UI
description: 在编写、审查或调试任何使用 MyGo（github.com/egoist/mygo）的 ZPad 代码前使用，包括 package ui、ui/editor、窗口、对话框、菜单、主题、字体、i18n、应用生命周期与文件 API。要求先查项目锁定版本的源码或官网文档，禁止凭记忆猜 API。
---

# MyGo 原生 UI

ZPad 是基于 MyGo 原生 UI 的编辑器。不要凭记忆猜 MyGo 的组件、函数或参数；先查再写。

## 查询顺序

1. **项目锁定版本的本地源码（最高优先级）**
   - 先看 `go.mod` 里的版本；当前是 `github.com/egoist/mygo v0.10.0`。
   - `MOD="$(go env GOMODCACHE)/github.com/egoist/mygo@v0.10.0"`
   - `go doc github.com/egoist/mygo/ui.Button`
   - `grep -rn "func Button(" "$MOD/ui" --include="*.go"`
   - 行为不确定时读实现和测试。
   - 本地模块才是实际编译的版本；官网文档可能对应更新版本。

2. **MyGo 官方文档（理解用法和示例）**
   - 索引：<https://mygo.run/llms.txt>
   - 全文：<https://mygo.egoist.dev/llms-full.txt>
   - 先在索引里找到页面，再抓取对应 `.md`。
   - 大页面用 `curl -sL <url> | grep -n`，不要整篇塞进上下文。
   - 不知道页面地址时先在 `llms.txt` 里搜，不要猜 URL。

3. **项目封装**
   - 主题 / 颜色：`internal/theme`
   - 应用状态 / 操作：`internal/app`
   - 设置：`internal/settings`
   - 编码：`internal/encoding`
   - 文件树：`internal/files`
   - 会话：`internal/session`
   - 用户文案：`locales` 生成函数

## 常用文档

| 领域 | 页面 |
| --- | --- |
| 应用生命周期 | `docs/app.md` |
| 窗口 | `docs/windows.md` |
| 布局 | `docs/ui/layout.md` |
| 样式 / 主题 | `docs/ui/styling.md` |
| 自定义控件 | `docs/ui/custom-widgets.md` |
| 代码编辑器 | `docs/ui/code-editor.md` |
| 菜单 | `docs/menus.md` |
| 对话框 / 桌面 API | `docs/native.md` |
| 国际化 | `docs/i18n.md` |
| 输入 / 滚动 | `docs/ui/input.md`、`docs/ui/scroll.md` |

某个具体控件：用 `llms.txt` 找 `docs/ui/<widget>.md`，或直接搜本地 `ui/api_gen.go`。

## 硬性规则

- 写调用前必须知道**确切的导出名和签名**。查不到就停下来说明，不要编一个看起来合理的 API。
- 不要依赖其它项目或其它 MyGo 版本的经验。
- 有项目封装时优先用封装，不直接调原始 API。
- UI 状态修改在主线程；文件读写和原生对话框放在 goroutine，完成后用 `win.Update` 回到主线程。
- `editor.Buffer` 只接受 UTF-8；文件编码转换统一走 `internal/encoding`。
- 字体必须在任何窗口创建前用 `ui.RegisterFont` 注册。
- 不引入 WebView，本项目只用 MyGo 原生 UI。
- 用户可见文案加到 `locales/en.json` 和 `locales/zh-Hans.json`，运行 `go tool mygo generate`，通过 `locales.X()` 使用，不要硬编码。
- 主题颜色放在 `internal/theme`，不要在 view 里散落硬编码色值。

## 常用命令

```sh
# 查精确 API
go doc github.com/egoist/mygo/ui.Button
go doc github.com/egoist/mygo/ui/editor.State

# 查源码
MOD="$(go env GOMODCACHE)/github.com/egoist/mygo@v0.10.0"
grep -rn "func Button(" "$MOD/ui" --include="*.go"
grep -rn "type State struct" "$MOD/ui/editor" --include="*.go"

# 查文档
curl -sL https://mygo.run/llms.txt | grep -n "code editor"
curl -sL https://mygo.egoist.dev/docs/ui/code-editor.md | head -80

# 验证
gofmt -w .
go test ./...
go vet ./...
go build ./...
go tool mygo build
```

## 修改前检查

- [ ] 已在本地源码或官方文档找到该符号。
- [ ] 已确认签名和所属包。
- [ ] 已检查 `internal/app`、`internal/theme`、`internal/encoding`、`locales` 是否已有封装。
- [ ] 已区分主线程与 goroutine。
- [ ] UI 行为准备了 `ui.Tester` 或单元测试。
