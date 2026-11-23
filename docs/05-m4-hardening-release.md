# 05 · M4 加固、文件关联与 macOS 发布

> 预估：1–2 天
> 前置：M3 已签收

## 1. 目标

把 ZPad 从“能用”做到“可以交付”：

- 大文件、二进制、异常编码、磁盘错误等边界情况可控。
- Finder 双击、拖拽、命令行打开都能工作。
- 产出可运行的 macOS `.app` / `.dmg`。
- 文档、截图、测试齐全。

## 2. 范围

### 做

- 大文件与二进制策略。
- 编码边界与往返校验。
- 保存失败、权限、文件被删除 / 改名等错误处理。
- 外部修改的基础检测（窗口获得焦点时检查 mtime）。
- 文件关联、Finder 打开、拖拽打开、命令行打开。
- macOS 打包、图标、版本信息。
- README、截图、最终文档。
- 全量测试与手工 QA。

### 不做

- 完整的文件监听与三方冲突合并（二期）。
- 热退出恢复未保存内容（二期）。
- 代码签名与公证（如需发布再单独处理）。
- Windows / Linux 打包。

## 3. 任务拆解

### 3.1 大文件与二进制

- [ ] 打开前 `os.Stat`：
  - `> 20 MB`：提示“文件较大，将以只读方式打开”，确认后 `state.ReadOnly = true`。
  - `> 50 万行`：同样只读提示。
- [ ] 读取放在 goroutine；打开期间显示“正在打开…”，避免 UI 卡死。
- [ ] 二进制检测：前 8 KB 含 `0x00` 即拒绝。
- [ ] 拒绝时给出文件名与原因，不加入文档列表。
- [ ] 只读文件在状态栏显示“只读”；尝试编辑时无反应或有提示。
- [ ] 大文件不参与会话恢复的滚动位置计算时可降级（可选）。

### 3.2 编码边界

- [ ] 测试夹具：
  - UTF-8 无 BOM / 有 BOM。
  - UTF-16 LE / BE 有 BOM。
  - UTF-16 LE / BE 无 BOM（启发式）。
  - GB18030 中文、生僻字、四字节序列。
  - 混合换行符。
  - 非法 UTF-8 字节。
- [ ] UTF-16 无 BOM 判断失败时，弹出编码选择对话框，列出候选并允许预览前几行。
- [ ] 保存前往返校验：

```go
encoded, err := encoding.Encode(text, enc)
if err == nil {
    back, err := encoding.Decode(encoded, enc)
    if err != nil || back != text {
        // 提示改用 UTF-8
    }
}
```

- [ ] BOM 选项：UTF-8 与 UTF-8 BOM 是两个独立选项，保存结果必须可验证。
- [ ] 重新以编码打开时：
  - 未修改：直接重读。
  - 已修改：提示“放弃未保存修改？”。
  - 失败：保持原文档不变，显示错误。
- [ ] 状态栏编码菜单的当前项有勾选标记。

### 3.3 错误处理

- [ ] 保存失败场景：
  - 目录只读 / 无权限。
  - 磁盘已满。
  - 目标文件被删除。
  - 目标文件被其他程序占用（Windows 后续）。
  - 临时文件创建失败。
- [ ] 失败时保留脏状态与编辑内容；错误信息包含路径与原因。
- [ ] 原子写入失败时清理临时文件，不留下垃圾。
- [ ] 文件在打开期间被删除：保存时提示“文件已不存在，是否重新创建？”。
- [ ] 文件在打开期间被外部修改：
  - 窗口获得焦点时比较 mtime / size。
  - 未修改：自动重新加载并 toast。
  - 已修改：提示“文件已被外部修改”，提供 重新加载 / 保留我的修改 / 另存为。
- [ ] 未捕获 panic 写入日志，尽量保持窗口可用。

### 3.4 文件关联与系统打开

- [ ] `mygo.json`：

```json
{
  "fileAssociations": [
    {
      "ext": ["txt", "md", "markdown", "html", "htm", "css", "json",
              "js", "mjs", "cjs", "ts", "tsx", "jsx", "xml",
              "yaml", "yml", "toml", "ini", "conf", "log", "csv"],
      "name": "Text Document",
      "mimeType": "text/plain"
    }
  ]
}
```

- [ ] `App.OnOpenFile`：
  - 窗口已就绪：在活动栏打开。
  - 窗口未就绪：缓存路径，首帧后打开。
  - 多个文件：全部打开，最后一个激活。
- [ ] 拖拽：
  - 文件：活动栏打开。
  - 文件夹：作为工作区打开。
  - 混合拖入：先打开文件夹，再依次打开文件。
  - 拖拽悬停高亮。
- [ ] 命令行：

```sh
ZPad.app/Contents/MacOS/ZPad /path/to/file.md
ZPad.app/Contents/MacOS/ZPad /path/to/folder
```

- [x] Windows / Linux 单实例锁与参数转发（`RequestSingleInstanceLock` + `OnSecondInstance`）。
- [ ] macOS Finder 双击 `.md` / `.txt` 验证关联生效（需要重新注册应用）。

### 3.5 macOS 打包

- [ ] 正式图标：1024×1024 PNG，替换 `resources/icon.png`。
- [ ] 版本信息：`0.1.0`，identifier `com.xiaoz.zpad`。
- [ ] 构建：

```sh
go tool mygo build
```

- [ ] 产物：
  - `build/darwin-arm64/ZPad.app`
  - `build/darwin-arm64/zpad-0.1.0-arm64.dmg`
- [ ] 在干净的用户环境（或新 macOS 用户）启动验证：
  - 首次启动无崩溃。
  - 设置与会话目录正确创建。
  - 文件关联可打开。
- [ ] 记录未签名应用的 Gatekeeper 提示与打开方式；签名 / 公证如需要，另开任务。
- [ ] Intel Mac 的 universal 构建在需要时再验证。

### 3.6 文档与截图

- [ ] 根目录 `README.md`：
  - 功能简介。
  - 截图（亮色 / 暗色）。
  - 环境要求。
  - 开发、测试、构建命令。
  - 快捷键表。
  - 已知限制。
- [ ] `docs/` 更新：
  - 各里程碑状态与签收日期。
  - 实际实现与方案的差异记录。
- [ ] `CHANGELOG.md` 记录 0.1.0。

### 3.7 测试与 QA

- [ ] 单元测试补充：
  - 大文件阈值判断。
  - 二进制检测。
  - 编码夹具全覆盖。
  - 保存失败路径（只读目录）。
  - 文件关联路径解析。
- [ ] 视图测试：
  - 只读状态显示。
  - 外部修改提示。
- [ ] 手工 QA 矩阵：

| 场景 | 操作 | 预期 |
| --- | --- | --- |
| UTF-8 | 打开、编辑、保存 | 内容与编码不变 |
| UTF-8 BOM | 打开、保存 | BOM 保留 |
| UTF-16 | 打开、保存 | 编码不变 |
| GB18030 | 打开中文、保存 | 编码不变 |
| 大文件 | 打开 20 MB 文件 | 只读、可滚动 |
| 二进制 | 打开 PNG | 拒绝并提示 |
| 拖拽 | 拖入文件 / 文件夹 | 正确打开 |
| Finder | 双击 `.md` | 用 ZPad 打开 |
| 关闭 | 3 个未保存文件 | 汇总提示 |
| 取消 | 关闭时取消 | 内容保留 |
| 暗色 | 强制暗色 | 菜单 / 对话框同步 |
| 语言 | 切换 English | 全部文案切换 |

- [ ] 性能检查：
  - 启动到可交互 < 1 秒。
  - 20 MB 文件打开 < 3 秒（M 系列 Mac）。
  - 输入延迟无可感卡顿。
  - 双栏 + 10 个标签内存占用可接受。

## 4. 交付物

- macOS `.app` 与 `.dmg`。
- 完整文件关联与拖拽打开。
- 边界情况处理与错误提示。
- README、CHANGELOG、截图。
- 全量测试与 QA 记录。

## 5. 验收清单

- [ ] 20 MB 文本文件以只读方式打开，滚动流畅。
- [ ] 二进制文件被拒绝，提示清晰。
- [ ] UTF-16 无 BOM 文件可通过手动选择编码正确打开。
- [ ] UTF-8 / BOM / UTF-16 / GB18030 夹具全部往返一致。
- [ ] 保存到只读目录失败时，内容不丢失，错误可读。
- [ ] 外部修改文件在窗口聚焦时被发现并提示。
- [ ] Finder 双击 `.md` / `.txt` 用 ZPad 打开。
- [ ] 拖拽文件 / 文件夹到窗口正确打开。
- [ ] 命令行传入文件 / 文件夹正确打开。
- [ ] `go tool mygo build` 产出 `.app` 与 `.dmg`。
- [ ] 干净环境首次启动正常，设置与会话目录创建正确。
- [ ] `go test ./...` 通过。
- [ ] README、CHANGELOG、截图完成。
- [ ] 全流程无 WebView。

## 6. 风险与备注

- 未签名应用在 macOS 上会被 Gatekeeper 拦截；正式分发需要 Apple Developer 账号与公证。
- 文件关联需要系统重新注册应用，开发期可能需手动清理 LaunchServices 缓存。
- 外部修改的完整监听与冲突合并放二期；本期只做窗口聚焦时的 mtime 检查。
- 大文件只读模式下，查找替换仍可用，但不允许编辑。

## 7. 实现记录

完成日期：2026-11-22。

已实现：

- 大文件只读：超过设置阈值（默认 20 MB）或 50 万行的文件，先提示再以只读打开；状态栏显示「只读」。
- 二进制文件（前 8 KB 含 NUL）拒绝打开并给出错误。
- 打开文件期间状态栏显示「正在打开…」，读取在 goroutine 中进行。
- 编码：UTF-8 / BOM / UTF-16 LE/BE / GB18030；UTF-16 无 BOM 启发式；保存前往返校验；状态栏编码菜单当前项打勾。
- 错误处理：保存失败保留脏状态与原内容；原子写临时文件失败清理；文件被删除时保存提示重新创建；只读目录/权限错误可见。
- 外部修改：窗口获得焦点时比较 mtime / size；未修改自动重载并 toast，有修改时选择重新加载 / 保留我的修改 / 另存为。
- 文件关联：`mygo.json` 声明常见文本扩展名，构建后的 Info.plist 含 `CFBundleDocumentTypes`。
- `App.OnOpenFile` 支持窗口未就绪缓存、单/多文件打开；拖拽文件夹先于文件处理；支持命令行。
- 单实例：`mygo.App.RequestSingleInstanceLock()` 在 `main` 开头拦住第二个进程；`OnSecondInstance` 把后续启动的文件 / 文件夹交给第一个实例并前置窗口（Windows / Linux；macOS 由系统派发）。参数由应用自行解析，因此不受 MyGo 只转发已登记扩展名的限制；正在读取的同一路径只读一次。
- 正式应用图标（1024×1024，代码生成），JetBrains Mono OFL 许可随应用资源提供。
- 关于对话框：窗口内主题绘制，显示应用图标、版本、AGPL-3.0（链接到仓库 LICENSE）与 GitHub 仓库；替换 macOS 系统默认的 About 项，Windows / Linux 从「帮助 → 关于 ZPad」进入。图标在启动时从 `PathResources/icon.png` 解码一次，缺失时回退为字母标；版本号由 `app.Version` 统一，`release_test.go` 校验它与 `mygo.json` 一致。
- macOS `.app` 与 `.dmg` 打包，README 截图 / 快捷键 / 发布说明、CHANGELOG。

已验证：

- `go build ./...`、`go test ./...`、`go vet ./...` 通过。
- `go tool mygo build` 通过：
  - `build/darwin-arm64/ZPad.app`（13.7 MB）
  - `build/darwin-arm64/zpad-0.1.0-arm64.dmg`（5.5 MB）
- Info.plist 已包含 `CFBundleDocumentTypes`（txt / md / html / css / json / js 等）。
- 真实系统打开验证：启动应用后 `open -a ZPad.app` 两个文件，退出后的 `session.json` 记录了这两个文件，说明多文件 `OnOpenFile` 生效。
- 单元测试覆盖：大文件按行数 / 大小只读、只读状态栏、只读目录保存失败、外部文件重载、GB18030 四字节、UTF-16 无 BOM、混合换行、文件关联配置、图标尺寸。

与方案的差异 / 待补：

- UTF-16 无 BOM 检测失败时没有独立的“预览并选择编码”对话框；当前通过状态栏「重新以…编码打开」手动选择，满足验收项。
- 外部修改只在窗口聚焦时检查，不是完整文件监听。
- 未实现未捕获 panic 的统一日志包装；依赖 MyGo 的崩溃处理。
- 未签名、未公证；Gatekeeper 拦截需右键打开或在系统设置允许。
- 未验证 Intel universal 构建与干净 macOS 用户环境，后续发布前补。
- 大文件只读模式下仍可查找替换和另存为，但不允许编辑。

发布前界面打磨（同日）：

- 全局滚动条改为 6px 圆角半透明拇指 + 10px 命中区，轨道透明；编辑器、侧栏、设置面板一致。
- 文件树选中改为全宽浅色（亮色 6% 黑、暗色 7% 白），不再使用饱和主色胶囊。
- 应用内自绘的按钮、菜单按钮、标签、标签关闭、新建标签、文件树行、设置下拉与开关显示手型光标；原生菜单保持系统光标。
- 文件树行改为整行自绘：缩进、折叠箭头、图标、文件名与左右空白都在同一个 `Row` 内，`ui.OutlineState.Prefix` 置空去掉 Outline 自带的缩进与箭头，行内任意位置单击都能打开文件 / 展开文件夹并同步选中；Option 点击行展开或收起其下全部文件夹。
- 标签栏隐藏横向滚动条；标签溢出时在末尾显示下拉按钮，列出该栏全部标签、当前项打勾，选择后切换并滚动到可见位置。
- 标签宽度设上限：文件名文本最多 180 DIP（标签整体约 240 DIP），短名保持自然宽度；超长名中间省略，同时保留开头与扩展名，悬停显示完整路径；脏点与关闭按钮不收缩。新增 `TestLongTabNameIsCapped`。
- 侧栏的路径提示改到行的右侧：`pathTooltip` 用 `ui.TooltipBase` + `AttachTo(ui.AnchorEnd, ui.AnchorStart)`，与行垂直居中对齐，不再压住列表和「打开的文件 / 工作区」栏头；「打开的文件」和工作区文件树共用，宽度上限 420 DIP，长路径按分隔符换行；配色沿用主题 tooltip 的反色（新增 `theme.Inverse`）。新增 `TestSidebarTipSitsBesideTheRow`、`TestInverseColors`。
- 「打开的文件」去掉右侧目录列：行内只显示文件名，名字占满整行、真放不下才中间省略（原来被硬限在 55%，`03-m2-split-session.md` 这类长名会被截断）；同名文件靠悬停的完整路径区分。删除 `documentDir` 与其单测，新增 `TestOpenFilesShowNamesOnly`。
- 设置面板改用 `ui.DialogBase` 直接绘制单层圆角面板，去掉 MyGo Modal 外层内边距与背景框；关闭按钮从标题栏移到「完成」左侧，Esc / 点击遮罩仍可关闭。

Windows / Linux 单实例（同日追加）：

- `main` 开头调用 `mygo.App.RequestSingleInstanceLock()`；第二个进程把命令行交给第一个实例后立即退出，不再出现第二个窗口。
- `OnSecondInstance` 里 `Restore` + `Focus` 已有窗口，再按应用自己的规则解析参数（跳过 `-` 开头、相对路径按第二个进程的工作目录解析），文件和文件夹都交给 `OpenPaths`；窗口未就绪时并入 `pendingOpen`。
- MyGo 的 `fileArgs` 只转发 `mygo.json` 中登记扩展名的参数，因此应用自行解析，`.go` / `.py` / 无扩展名等用「打开方式」选 ZPad 时也能打开。
- `openPaths` 增加「正在读取」集合（`beginLoading` / `endLoading`），同一路径来自 `OnOpenFile` 和第二实例时只读一次；顺带消除启动时 `pendingOpen` 与 `os.Args` 重复读取。
- Windows 安装包只把 ZPad 写入 `.ext\OpenWithProgids` 与 ProgID，默认应用仍需用户右键「打开方式 → 始终使用此应用」（系统限制，非应用问题）。
- 待验证：真实 Windows 上的安装、首次设为默认、已运行时双击第二个文件只保留一个 `ZPad.exe`、退出后 `settings.json` / 会话不被旧实例覆盖。

文件关联扩展（同日追加）：

- `mygo.json` 拆成 `Text Document`（文档 / 数据 / 标记，40 个）与 `Source Code`（源码，46 个）两组，共 86 个扩展名；`internal/app/textExtensions`（打开对话框的文本筛选器）同步，`release_test.go` 校验两份列表完全一致、同一扩展名只出现在一个分组。
- 覆盖 MyGo 编辑器能高亮的常见扩展名：文档类含 md / mdx / html / xhtml / css / scss / less / jsonc / yaml / toml / ini / properties / editorconfig / diff；源码类含 go / py / rs / c / cpp / h / java / kt / scala / swift / dart / cs / rb / php / lua / zig / sh / sql / proto / graphql / tf / nix / erb / jinja 等。
- 有意不关联：`.command`（双击会执行脚本）、`.plist` / `.svg` / `.xib` / `.storyboard` / `.csproj`（系统或 IDE 资产）、`.ino`、`.m` / `.mm`（歧义多）。无扩展名文件（Dockerfile / Makefile / README）只按文件名高亮，系统关联只认扩展名。
- macOS 的 `CFBundleDocumentTypes` 由 MyGo 统一写 `LSHandlerRank: Default`；Windows 安装包写入 `HKCU\Software\Classes` 的 ProgID 与 `OpenWithProgids`，默认程序仍由用户选择。
- 已验证：`go tool mygo build` 产出的 Info.plist 含两组 `CFBundleDocumentTypes`，共 86 个扩展名、无重复，角色为 Editor、`LSHandlerRank` 为 Default；Finder「打开方式」可列出 ZPad。
- 侧栏改为「打开的文件 / 工作区」双 Tab：打开文件列表汇总两栏文档、去重，行内只显示文件名；打开文件夹或恢复工作区时自动切到工作区。
- 文件树改为单击打开文件、单击展开/收起文件夹，去掉双击处理。
- 标签可拖到另一栏，源栏最后一个标签移走后自动收起分屏；同一文件已在目标栏时直接激活。
- 侧栏工作区里的文件可拖到任意一栏打开；某栏关闭最后一个标签后自动收起分屏，旧会话里的空栏也会收敛。
- 文件树右键菜单：在副窗口打开、重命名、删除；重命名同步打开标签路径、窗口标题和会话，删除前拦截未保存修改，删除为永久删除。
- 「打开的文件」列表对已保存文件提供同样的右键操作；全部弹窗菜单改为窗口内主题绘制，顶部菜单栏保持原生。
- 应用内菜单项悬停显示手型光标；MyGo 内部的编辑器右键和输入框编辑菜单仍是箭头。
- 修复文件树单击打开后选中行不跟随的问题：行内容用 `Clicked()` 打开文件，并手动更新 `treeSelected`。
- 标签右键菜单：关闭当前 / 左侧 / 右侧、移动到主窗口 / 副窗口、复制到主窗口 / 副窗口；右键不切换活动标签；未分屏且只有一个标签时「移动到副窗口」禁用，避免空栏收起分屏；复制共享文档但使用独立的光标 / 滚动状态。
- 批量关闭（左侧 / 右侧）遇到多个未保存文档时只弹一次「保存全部 / 不保存 / 取消」，避免逐个弹框；关闭后按现有规则在栏空时收起分屏。
- 修复「关闭左侧 / 右侧标签」只关掉一部分、关闭左侧还会误删右侧标签：菜单把 `p.tabs` 的子切片交给 `closeTabs`，而 `removeTab` 的 `append` 会在同一个底层数组上左移元素，遍历时指针每删一个就跳两格；`closeTabs` 开头改为对入参做快照（`append([]*Tab(nil), tabs...)`）。`tabmenu_test.go` 新增 5 个标签的左右批量关闭用例（直接调用 `closeTabs` 并传入子切片，因为 tester 没有窗口、移除会在帧内立即执行）。
- 新增 `internal/app/tabmenu_test.go`；`go test ./...`、`go vet ./...`、`go build ./...` 通过。
- 点击分屏按钮时按内容区（窗口减去侧栏和 1px 分隔线）50/50 拆分。
- 新增 `theme.FileTree`、`locales.allTabs`、`internal/app/tabbar_test.go`，以及滚动条 / 树主题单测。
- 内存优化（实测 Activity Monitor 物理占用）：`main.go` 把 Go 堆目标设为 `debug.SetGCPercent(50)`，堆高水位从 24–28 MB 降到 14 MB；About 的 1024×1024 图标改为首次打开对话框时解码。同一会话（7+1 个文档、最大化 1920×995、焦点移走后稳定）空载从 80–87 MB 降到 68–73 MB，新启动空载 62 MB 不变。
- 大文件内存：`Document` 原先把 `os.ReadFile` 的整份字节存在 `Raw` 里，但全项目没有读取处（外部修改检测用 `Size` + `ModTime`，保存从 `Buffer` 重新编码，重载与换编码会重新读盘）。删掉字段和 3 处赋值后，打开期间内存里只剩 rope 一份文本：20 万行 / 13.8 MB 的文档存活堆 44 MB → 20 MB，119 MB / 15 万行的文件 355 MB → 178 MB，约 1.1 倍文件大小；`GODEBUG=gctrace=1` 核对过，行为不变。
- 诊断日志（`internal/app/log.go`）：写到用户数据目录的 `zpad.log`，`log.SetOutput(io.MultiWriter(文件, stderr))`，文件排在前面，Windows 上 stderr 无效时也能落盘；启动时超过 1 MB 轮转成 `zpad.log.1`，最多两个文件；`ZPAD_LOG=off` 关闭，`ZPAD_LOG=<路径>` 换位置。记录启动版本 / 平台、窗口大小、打开 / 重载 / 保存的耗时与大小、关闭高亮的判定，以及 MyGo 自己打印的渲染器日志——Windows 打包版是 `-H=windowsgui`，没有控制台，`mygo: DirectComposition draws with WARP` 这类信息只能从这里看到。菜单「查看日志」（macOS 应用菜单、其他平台 Help）用编辑器打开它。
- 慢帧统计：`app.NoteFrame` 包在 `ui.View` 外，每帧只做一次 `time.Since` 和定长数组写入，**0 次分配**（`TestNoteFrameAllocatesNothing` 用 `testing.AllocsPerRun` 守住）；只有出现 > 12 ms 的帧才每 10 秒写一行 `frames 10s: n=… p50=… p95=… max=… slow=…`，正常运行时日志保持安静（`TestFastFramesStayQuiet`）。前 10 帧不计，避免首帧塑形的噪声。
- 大文件不高亮：`Document.setLanguage` 在行数超过 `maxLines`（50 万，与只读阈值同一个数）时只把语言记在 `Document.Language` 上，`Buffer` 设为 nil；打开、两处重载、保存、保存全部、重命名六处统一走它，避免保存或改名后高亮「复活」。状态栏用 `locales.LanguageNoHighlight` 显示「Go · 未高亮」，普通文件不变。
- 试过关闭全部标签后 `debug.FreeOSMemory` + `SetMemoryLimit`，实测 `HeapReleased` 只增加约 3 MB、Activity Monitor 无变化，已去掉，保持改动最小。
- `go test ./...`、`go vet ./...`、`go build ./...` 通过。

错误提示改为自动消失的 toast（同日追加）：

- 问题：打开 / 拖入不支持的文件时，失败写进状态栏 `errMsg`，而 `errMsg` 只在成功打开文件或打开文件夹时清零，红色提示会永远留在左下角，并顶掉状态栏的文件路径。
- 现在动作失败一律走 `SetErrorToast(title, detail)`：窗口底部显示两行 toast——标题「无法打开 alipay_500.jpg」、说明「这是二进制文件，无法作为文本编辑」，6 秒后自动消失，指针悬停暂停计时，Esc 可关闭，同标题重复触发只刷新不堆叠；屏幕阅读器同时播报。
- 一次拖入多个文件时由 `reportOpenFailures` 汇总：单个文件报文件名与原因，多个文件报「有 3 个文件无法打开」加去重后的原因（最多 3 条，`locales.reasonSeparator` 连接），不再一文件一条。
- toast 队列由 `pendingToasts []ui.Toast` 保存，视图每帧 `AddToast` 后清空；`SetToast`（删除、外部重载、设置恢复）与错误 toast 共用队列，同一帧内的多条不会互相覆盖。
- 启动期提示（会话缺失文件、设置读取 / 初始化失败、字体回退）也改为 toast，不再在状态栏留红字；`SetError` 只保留给「文档还开着、需要用户处理」的保存 / 关闭失败。
- 保存失败原本同时写状态栏和弹对话框，现在只弹对话框，避免同一件事说两遍。
- `unknownEncoding` 文案改为「无法识别文件编码，无法作为文本打开。」：文件没打开时状态栏没有文档，原来「请在状态栏选择编码重新打开」指不到任何控件。
- 新增 `internal/app/toast_test.go`：单个失败显示标题与原因并在 6 秒后消失、批量失败只出一条并带计数与原因、`fail` 不再写状态栏；另用 `ui.Tester` 截渲亮 / 暗两套人工确认 toast 贴着状态栏上沿、不遮挡编码与换行菜单。
- `go test ./...`、`go vet ./...`、`go build ./...` 通过。

需要人工确认：

- [ ] 20 MB / 50 万行文件以只读打开，滚动流畅
- [ ] 二进制文件被拒绝
- [ ] 拖入二进制文件后错误 toast 约 6 秒自动消失，悬停暂停，Esc 关闭；一次拖入多个不支持文件只出一条汇总
- [ ] UTF-16 无 BOM 用状态栏手动编码打开
- [ ] UTF-8 / BOM / UTF-16 / GB18030 打开保存往返
- [ ] 只读目录保存失败，内容不丢
- [ ] 外部修改后聚焦窗口，三种处理方式
- [ ] Finder 双击 `.md` / `.txt` 用 ZPad 打开
- [ ] 拖拽文件 / 文件夹、混合拖入
- [ ] 命令行打开文件 / 文件夹
- [ ] `.app` / `.dmg` 在干净环境启动
- [ ] 未签名应用的 Gatekeeper 提示

## 8. 签收

| 项目 | 内容 |
| --- | --- |
| 状态 | 待签收 |
| 签收人 | |
| 签收日期 | |
| 实际用时 | 2026-11-22 · 约 2 小时 |
| 备注 | 代码、打包与自动化检查已完成；等待 0.1.0 发布验收 |

- [ ] 验收清单全部通过
- [ ] 相关文档已更新
- [ ] 同意发布 0.1.0

签收意见：

```text

```
