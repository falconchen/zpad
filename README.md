# ZPad

基于 [MyGo](https://mygo.run/) 原生 UI 的轻量级文本 / 代码编辑器。
单窗口、左右双栏、每栏多标签，不启动 WebView，界面全部由 Go 绘制。

当前进度：**M2 左右双栏与会话恢复**。已支持打开文件 / 文件夹、侧栏文件树、多标签、语法高亮、保存与编码切换、左右分屏、会话恢复，以及关闭 / 退出时的未保存汇总提示。

## 环境

- Go **1.27.4**（官网 tarball 安装，`/usr/local/go`）
- macOS（一期只验收 macOS；Windows / Linux 后续适配）
- 原生 UI 无需 Bun、Node、WebView

## 开发

```sh
go tool mygo dev     # 启动开发版，改 .go 文件自动重启
go test ./...        # 不打开窗口的视图测试与单元测试
go tool mygo build   # 打包 macOS 应用
```

首次运行 `go tool mygo dev` 时，Go 会按 `go.mod` 准备 MyGo CLI 与依赖。

## 目录

```
.
├── main.go                 # 应用入口：设置、窗口、主题源
├── mygo.json               # 应用名称 / 标识 / 版本
├── go.mod                  # Go 1.27.4 + MyGo v0.10.0
├── docs/                   # 需求与里程碑文档
├── locales/                # 中英文 catalogs（M3）
├── resources/              # 应用图标等资源
└── internal/
    ├── app/                # 应用状态、文档、标签、视图、菜单
    ├── atomicfile/         # 原子写文件
    ├── caseconv/           # 大小写转换
    ├── encoding/           # 编码检测 / 解码 / 编码
    ├── files/              # 工作区目录树
    ├── session/            # 会话读写（M2 接入）
    ├── settings/           # 设置模型与持久化
    └── theme/              # 亮色 / 暗色调色板与编辑器主题
```

## 文档

- [需求与里程碑索引](./docs/README.md)
- [最终需求](./docs/00-requirements.md)
- [M0 项目骨架](./docs/01-m0-skeleton.md)
- [M1 单栏编辑闭环](./docs/02-m1-editor-core.md)
- [M2 左右双栏与会话恢复](./docs/03-m2-split-session.md)
- [M3 视觉、设置、字体与中英双语](./docs/04-m3-appearance-settings.md)
- [M4 加固、文件关联与 macOS 发布](./docs/05-m4-hardening-release.md)

## 已实现

- 打开文件、打开文件夹、新建文件、拖拽打开。
- 侧栏文件树（默认关闭，打开文件夹时展开），隐藏文件开关、刷新。
- 多标签、未保存圆点、关闭标签提示。
- 语法高亮、行号、多光标、折叠、查找替换、跳转行。
- 保存 / 另存为，原子写入。
- UTF-8 / UTF-8 BOM / UTF-16 LE/BE / GB18030 检测与切换。
- LF / CRLF / CR 显示与转换。
- 大小写转换（大写、小写、首字母大写、反转，多选区）。
- 亮色 / 暗色 / 跟随系统，编辑器 token 配色。
- 状态栏：路径、语言、编码、换行、行列、选区、大小、未保存标记。
- 原生菜单与快捷键。
- 左右双栏：每栏独立标签，同一文件跨栏共享 Buffer，活动栏跟随焦点。
- 会话恢复：工作区、分屏、宽度、标签、活动标签、滚动位置、侧栏。
- 关闭 / 退出未保存汇总：保存全部 / 不保存 / 取消。

## 当前限制

- 设置面板、中英双语、内置字体在 M3。
- 全局搜索、Markdown 预览、外部修改监听在二期。
- 超过 20 MB 的文件当前直接拒绝；M4 改为只读打开。
- 标签中键关闭、拖拽排序未实现。

## 许可证

AGPL-3.0，见 [LICENSE](./LICENSE)。
