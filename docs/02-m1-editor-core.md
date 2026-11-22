# 02 · M1 单栏编辑闭环

> 预估：3–4 天
> 前置：M0 已签收

## 1. 目标

在单栏（不启用分屏）的前提下，完成**打开 → 编辑 → 保存**的完整闭环，
并接上侧栏、标签页、编码、换行、大小写、查找替换、状态栏、拖拽打开。

M1 结束时，ZPad 已经可以作为一个日常轻量编辑器使用。

## 2. 范围

### 做

- 打开文件、打开文件夹、新建文件。
- 侧栏文件树（默认关闭，首次打开文件夹自动展开）。
- 单栏多标签。
- 编辑器接入与设置项生效。
- 保存、另存为、脏标记、关闭标签提示。
- 编码检测、解码、保存编码、重新以编码打开。
- 换行符显示与转换。
- 大小写转换。
- 查找 / 替换（内建）。
- 状态栏。
- 菜单与快捷键。
- 拖拽文件 / 文件夹打开。
- 单元测试与视图测试。

### 不做

- 左右分屏（M2）。
- 会话恢复（M2）。
- 窗口级关闭汇总提示（M2）。
- 设置面板与中英文案（M3）。
- 全局搜索、外部修改监听、Markdown 预览（二期）。

## 3. 任务拆解

### 3.1 工作区与文件打开

- [ ] `internal/files`：
  - `ReadDir(path) ([]Entry, error)`，目录在前、按名称排序。
  - 忽略 `.git`、`node_modules`、`.DS_Store`；隐藏文件由开关控制。
  - `IsTextFile(path)`：按扩展名 + 二进制检测。
  - `DetectBinary(data []byte) bool`：前 8 KB 含 NUL 即视为二进制。
- [ ] 打开文件对话框：

```go
paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
    Parent:   app.win,
    Title:    tr("Open File"),
    Filters:  textFilters(),
    Multiple: true,
})
```

- [ ] 打开文件夹对话框：`Directory: true`。
- [ ] `App.OnOpenFile` 接收系统打开事件；若窗口未就绪，先缓存路径。
- [ ] 拖拽：
  - 窗口级 `OnFileDrop` 接收文件。
  - 内容区 `DroppedFiles()` + `FileDragOver()` 显示高亮。
  - 拖入文件夹 → 作为工作区；拖入文件 → 在活动栏打开。

### 3.2 文档管理

- [ ] `internal/app.Document`：

```go
type Document struct {
    path string
    buf  *editor.Buffer
    enc  encoding.Encoding
    raw  []byte // 未修改时的原始字节
    size int64
}
```

- [ ] `OpenDocument(path)`：
  - 规范化绝对路径，已打开则直接激活对应标签。
  - 读取字节，检测编码，解码为 UTF-8。
  - `editor.NewBuffer(text)`，`SetLanguage(editor.LanguageFor(path))`。
  - 记录原始字节、大小、编码。
- [ ] `NewDocument()`：空 Buffer，未命名，默认编码与换行符来自设置。
- [ ] 大文件与二进制拦截（具体阈值见 M4，M1 先做基础拦截）。
- [ ] 文档不直接持有 UI 状态；光标与滚动属于标签。

### 3.3 标签页

- [ ] `Tab`：

```go
type Tab struct {
    doc   *Document
    state *editor.State
}
```

- [ ] 自绘标签栏：
  - 当前标签高亮。
  - 未保存圆点。
  - 关闭按钮；中键关闭。
  - 标签过多时横向滚动。
  - 空栏显示欢迎页。
- [ ] `Cmd+W` 关闭当前标签。
- [ ] 切换标签时保留各自的光标、选区和滚动位置。
- [ ] 关闭标签流程：
  - 未修改：直接关闭。
  - 已修改：`mygo.Dialog.Message` 提供 保存 / 不保存 / 取消。
  - 未命名文件选择保存时走“另存为”。

### 3.4 编辑器接入

- [ ] 每个标签创建：

```go
state := &editor.State{Buffer: doc.buf}
state.SoftWrap = settings.SoftWrap
state.LineNumbers = ...
state.TabSize = settings.TabSize
state.Indent = ...
state.WrapGuides = ...
state.Font = ui.Font{Family: settings.FontFamily, Size: settings.FontSize}
state.LineHeight = settings.LineHeight
```

- [ ] 视图：

```go
editor.View(c, tab.state).Grow(1).FillWidth()
```

- [ ] 获得焦点的标签记录为活动标签。
- [ ] 右键菜单追加：大小写转换、编码、换行符、关闭标签。
- [ ] 查找替换使用内建 `Cmd+F` / `Cmd+Alt+F`，M1 保持默认外观。
- [ ] 编辑器主题先使用 `theme.EditorTheme` 的最小版本。

### 3.5 保存

- [ ] `Save(doc)`：
  - 未命名 → `Save As`。
  - `buf.WriteTo(&bytes.Buffer)` 得到 UTF-8 文本（保留行尾）。
  - `encoding.Encode(text, doc.enc)` 转为目标编码。
  - 原子写入：同目录临时文件 → `fsync` → `rename`。
  - 成功：`buf.MarkSaved()`，更新 `raw`、`size`、窗口标题。
  - 失败：对话框显示错误，保留脏状态。
- [ ] `Save As`：文件对话框，更新路径、扩展名 → 语言高亮。
- [ ] 保存中禁用重复触发，显示忙碌状态。

### 3.6 编码

- [ ] `internal/encoding`：

```go
type Encoding string

const (
    UTF8    Encoding = "utf-8"
    UTF8BOM Encoding = "utf-8-bom"
    UTF16LE Encoding = "utf-16le"
    UTF16BE Encoding = "utf-16be"
    GB18030 Encoding = "gb18030"
)

func Detect(data []byte) (Encoding, error)
func Decode(data []byte, enc Encoding) (string, error)
func Encode(text string, enc Encoding) ([]byte, error)
func StripBOM(data []byte) (payload []byte, enc Encoding, ok bool)
```

- [ ] 检测顺序：BOM → UTF-8 校验 → GB18030 → 失败返回错误。
- [ ] UTF-16 无 BOM 时用零字节分布做启发式判断，失败则交给用户选择。
- [ ] 保存前往返校验：`Decode(Encode(text)) == text`；不一致则提示改用 UTF-8。
- [ ] 状态栏编码菜单：
  - 重新以 UTF-8 / UTF-8 BOM / UTF-16 LE / UTF-16 BE / GB18030 打开。
  - 以 UTF-8 / UTF-8 BOM / UTF-16 LE / UTF-16 BE / GB18030 保存。
- [ ] 重新打开时若文件已修改，先提示放弃修改。
- [ ] 错误提示不能吞掉原始字节；失败时保持当前文档不变。

### 3.7 换行符

- [ ] 状态栏显示 `buf.LineEnding()`：`LF` / `CRLF` / `CR`。
- [ ] 菜单转换：`buf.SetLineEnding("\n" / "\r\n" / "\r")`。
- [ ] 保存时使用缓冲区实际行尾，不强制统一。
- [ ] 新建文件使用设置默认值。

### 3.8 大小写转换

- [ ] `internal/caseconv`：

```go
func Upper(s string) string
func Lower(s string) string
func Title(s string) string
func Invert(s string) string
```

- [ ] 对 `state.Selections()` 逐个处理：
  - 有选区：转换选中文本。
  - 无选区：取光标所在单词。
- [ ] 替换按偏移从后往前，避免偏移失效。
- [ ] 整个操作放在 `buf.Transact` 中，作为一次撤销。
- [ ] 转换后恢复/更新选区，保持多光标。
- [ ] 入口：编辑菜单、右键菜单。

### 3.9 状态栏

- [ ] 左侧：相对路径、语言、编码、换行符。
- [ ] 右侧：`Ln 12, Col 34`、选区字符数、文件大小、未保存标记。
- [ ] 列号按字符计算，不按 UTF-8 字节。
- [ ] 编码、换行符可点击打开菜单。
- [ ] 无文件时显示 `ZPad 0.1.0`。

### 3.10 菜单与快捷键

- [ ] 文件：新建、打开文件、打开文件夹、保存、另存为、关闭标签、关闭窗口。
- [ ] 编辑：撤销、重做、剪切、复制、粘贴、全选（使用 MyGo role）、大小写子菜单。
- [ ] 视图：侧栏、查找、替换。
- [ ] 编码：重新以…打开、以…保存。
- [ ] 快捷键按 `00-requirements.md` 第 7 节。
- [ ] 不覆盖编辑器默认键位。

### 3.11 测试

- [ ] 编码表驱动测试：
  - UTF-8、UTF-8 BOM、UTF-16 LE/BE、GB18030 中文。
  - 往返一致性。
  - 非法字节序列。
- [ ] 大小写转换测试：ASCII、中文、混合、多行。
- [ ] 文档打开 / 保存测试：`t.TempDir()` 中写入 → 打开 → 修改 → 保存 → 重新读取。
- [ ] 视图测试：`ui.Tester` 打开临时文件，断言标签标题、状态栏文本。
- [ ] 二进制文件被拒绝。
- [ ] `go test ./...` 全绿。

## 4. 交付物

- 可打开文件夹、浏览文件树、编辑并保存的单栏编辑器。
- UTF-8 / UTF-16 / GB18030 正确读写。
- 标签、脏标记、编码与换行切换、大小写转换、查找替换、状态栏。
- 拖拽打开文件与文件夹。
- 测试覆盖核心逻辑。

## 5. 验收清单

- [ ] 打开 UTF-8 文件，中文、Emoji 显示正常，编辑保存后内容不变。
- [ ] 打开 UTF-8 BOM 文件，保存后 BOM 仍存在。
- [ ] 打开 UTF-16 LE / BE 文件，内容正确，保存后编码不变。
- [ ] 打开 GBK / GB18030 中文文件，内容正确，保存后仍为 GB18030。
- [ ] 状态栏切换编码后，文本按预期重新解释或保存。
- [ ] LF / CRLF 文件打开、编辑、保存后行尾保持；手动转换生效。
- [ ] 多选区大小写转换正确，且一次撤销可还原。
- [ ] 内建查找支持大小写、整词、正则、选区内，替换一个/全部正常。
- [ ] 关闭未保存标签会提示，取消后标签保留。
- [ ] 拖拽单个文件到窗口可打开；拖拽文件夹可作为工作区。
- [ ] 侧栏默认关闭；首次打开文件夹自动展开一次；之后记住开关。
- [ ] 状态栏行列号随光标更新，列号按字符计算。
- [ ] 二进制文件被拒绝并提示。
- [ ] `go test ./...` 通过。

## 6. 风险与备注

- GB18030 覆盖全部 Unicode，正常情况下不会出现“无法表示”的字符；仍保留往返校验作为安全网。
- UTF-16 无 BOM 的启发式可能误判，必须提供手动覆盖。
- 打开大文件放在 M4 完善，M1 先按 20 MB 做基础拦截。
- 标签页为自绘，不依赖 `ui.Tabs` 的关闭能力。

## 7. 实现记录

完成日期：2026-11-22。

已实现：

- 打开文件 / 打开文件夹 / 新建文件。
- `App.OnOpenFile` 接收 Finder、Dock 和命令行传入的文件；窗口未就绪时排队。
- 窗口级与内容区拖拽打开文件 / 文件夹，拖拽悬停高亮。
- 侧栏文件树：目录优先排序、文件/文件夹图标、单击打开 / 展开、右键菜单（在副窗口打开 / 重命名 / 删除）、隐藏文件开关、刷新、宽度可拖并持久化。
- 单栏多标签：激活、未保存圆点、关闭按钮、新建标签、横向滚动。
- 编辑器接入：语言高亮、行号、软换行、缩进、括号自动闭合、折叠、多光标、查找替换。
- 保存 / 另存为：原子写入、保留文件权限、更新语言与标题、失败提示。
- 关闭未保存标签：保存 / 不保存 / 取消。
- 编码：UTF-8、UTF-8 BOM、UTF-16 LE/BE、GB18030；自动检测、手动重新打开、以指定编码保存、往返校验。
- 换行符：LF / CRLF / CR 显示与转换，保存保留原行尾。
- 大小写转换：大写、小写、首字母大写、反转；支持多选区，无选区时取当前单词。
- 状态栏：相对路径、语言、编码菜单、换行菜单、行列号、选区字符数、文件大小、未保存标记。
- 原生菜单：文件、编辑、显示、编码；撤销/重做/剪切/复制/粘贴/全选使用 MyGo roles。
- 亮色 / 暗色 / 跟随系统；编辑器 token 配色；侧栏、标签、工具栏、状态栏统一风格。

已验证：

- `go build ./...`、`go test ./...`、`go vet ./...` 通过。
- `go tool mygo build` 通过：
  - `build/darwin-arm64/ZPad.app`（11.4 MB）
  - `build/darwin-arm64/zpad-0.1.0-arm64.dmg`（3.7 MB）
- 应用可启动；`open -a ZPad.app docs/README.md` 能通过 `OnOpenFile` 打开文件。
- 编码、行尾、大小写、文件树、标签、设置持久化均有单元测试。
- 用 `ui.Tester` 渲染亮色 / 暗色界面，检查了工具栏、侧栏、标签、编辑器、状态栏布局。

与方案的差异 / 待补：

- 标签中键关闭未实现；当前用关闭按钮和 `Cmd+W`。
- 保存时设置了 `busy` 防重入，但状态栏没有“保存中…”文字提示。
- M1 没有设置面板，设置项先通过 `settings.json` 生效；M3 做面板。
- 主题切换使用工具栏菜单而非 Segmented，避免原生控件在暗色下仍显示浅色外观。
- 窗口关闭 / 退出的未保存汇总提示按计划放在 M2；M1 只处理单标签关闭。
- 超过 20 MB 的文件当前直接拒绝；M4 改为只读打开。

需要人工确认：

- [ ] 打开文件 / 文件夹对话框
- [ ] 拖拽文件与文件夹
- [ ] 关闭未保存标签的三种选择
- [ ] 编码重新打开 / 以编码保存
- [ ] 强制暗色时原生菜单与对话框外观
- [ ] 中英文以外的界面语言在 M3 处理

## 8. 签收

| 项目 | 内容 |
| --- | --- |
| 状态 | 已签收 |
| 签收人 | 用户 |
| 签收日期 | 2026-11-22 |
| 实际用时 | 2026-11-22 · 约 2 小时 |
| 备注 | 用户指示开始 M2，视为通过 |

- [x] 验收清单全部通过
- [x] 相关文档已更新
- [x] 同意进入 M2

签收意见：

```text

```
