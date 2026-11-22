# ZPad 开发文档

ZPad 是一个基于 [MyGo](https://mygo.run/) 原生 UI 的轻量级文本 / 代码编辑器：
单窗口、左右双栏、每栏多标签，不启动 WebView，全部界面由 Go 绘制。

当前阶段：**M4 已实现，等待 0.1.0 签收**。

## 文档索引

| 文档 | 内容 |
| --- | --- |
| [00-requirements.md](./00-requirements.md) | 最终需求、范围、技术约束、交互与验收总则 |
| [01-m0-skeleton.md](./01-m0-skeleton.md) | M0 项目骨架与基础设施 |
| [02-m1-editor-core.md](./02-m1-editor-core.md) | M1 单栏编辑闭环 |
| [03-m2-split-session.md](./03-m2-split-session.md) | M2 左右双栏与会话恢复 |
| [04-m3-appearance-settings.md](./04-m3-appearance-settings.md) | M3 视觉、设置、字体与中英双语 |
| [05-m4-hardening-release.md](./05-m4-hardening-release.md) | M4 加固、文件关联与 macOS 发布 |

## 里程碑总览

| 阶段 | 主题 | 预估 | 状态 | 签收日期 |
| --- | --- | --- | --- | --- |
| M0 | 项目骨架与基础设施 | 0.5–1 天 | 已签收 | 2026-11-22 |
| M1 | 单栏编辑闭环 | 3–4 天 | 已签收 | 2026-11-22 |
| M2 | 左右双栏与会话恢复 | 1–2 天 | 已签收 | 2026-11-22 |
| M3 | 视觉、设置、字体与中英双语 | 2 天 | 已签收 | 2026-11-22 |
| M4 | 加固、文件关联与 macOS 发布 | 1–2 天 | 待签收 | — |

总计约 8–11 个工作日。

## 已确认决策

| 项目 | 决策 |
| --- | --- |
| 应用名 | **ZPad** |
| 应用标识（Bundle ID） | **`com.xiaoz.zpad`** |
| 初始版本 | `0.1.0` |
| 平台优先级 | macOS 先跑通 → Windows（后续）→ Linux（最后） |
| 渲染方式 | MyGo 原生 UI，不启动 WebView |
| 窗口模型 | 单窗口 + 左右双栏 + 每栏多标签 |
| 侧栏 | 树形目录，默认关闭；首次打开文件夹时自动展开一次 |
| 保存方式 | 手动保存 + 未保存标记 |
| 关闭未保存 | 汇总提示：保存全部 / 不保存 / 取消 |
| 编码 | UTF-8 / UTF-8 BOM / UTF-16 LE / UTF-16 BE / GB18030 |
| 查找替换 | MyGo 内建 FindBar（当前文件），只重绘样式 |
| 全局搜索 | 二期 |
| 预览 | 一期不做；Markdown 预览放二期 |
| 主题 | 亮色 / 暗色 / 跟随系统 |
| 界面语言 | 中英双语，默认跟随系统，可手动切换 |
| 字体 | 默认系统等宽；字号可调；内置 JetBrains Mono，可选系统等宽字体 |
| 文件关联 | 支持常见文本扩展名 |
| 拖拽打开 | 支持拖入文件或文件夹 |
| 许可证 | 沿用仓库现有 AGPL-3.0 |

## 签收流程

1. 每个里程碑完成后，把本页状态改为「待签收」，并在对应文档的「签收」一节补充实际完成情况。
2. 由你按该文档的验收清单逐项确认，填写签收日期与备注。
3. 签收通过后，状态改为「已签收」，再进入下一阶段；未通过则记录问题并继续修正。

## 环境

- Go：**1.27.4**（官网 tarball 安装，`/usr/local/go`，`GOROOT=/usr/local/go`）
- MyGo：`github.com/egoist/mygo`，项目内用 `go tool mygo` 固定版本
- 开发命令：`go tool mygo dev`
- 测试命令：`go test ./...`
- 构建命令：`go tool mygo build`

> 安装方式：下载 `https://go.dev/dl/go1.27.4.darwin-arm64.tar.gz`，核对官方 SHA-256 后解压到 `/usr/local/go`。Homebrew 的 Go 已移除，旧版 Go 1.23.2 已被替换。
