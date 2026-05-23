# 项目政策

本文件把长期项目规则放进公开文档，而不是只留在私有计划里。

## 范围

`goost` 是一组小型 Go 工具包集合。它不应该变成 web framework、配置框架、日志框架、ORM、依赖注入容器或应用运行时。

可以接受的新增能力应当可复用、依赖克制，并且不需要框架手册就能解释清楚。重依赖集成应放在 nested modules 中。demo-only 和 benchmark-only 依赖必须留在 root module 之外。

## 新增准入

新包或 exported API 进入 root module 前，应同时满足这些条件：

- 用一句话就能说明用途，而且不依赖某个具体应用；
- 至少有两个合理用户或包会从这个 API 受益；
- 标准库或现有 `goost` 包不能足够自然地覆盖这个需求；
- API 可以用一小段 README 和一个可编译 example 说明清楚；
- 对 root module 的依赖影响可以接受。

如果新增能力需要框架、服务 SDK、重 benchmark 工具或 demo-only 依赖，应放在 nested module、benchmark module 或 example module 中，而不是 root module。

## 命名

Package 和目录名应描述可复用边界，而不是某个调用点。`httpx`、`zapctx`、`slogctx`、`ttlmap`、`defaultmap` 这类短名称可以保留，但 README 表格和 package docs 必须清楚限定边界。

v1.0 之前应审查含义模糊、过窄或容易与 Go 常用术语冲突的名称。只有长期清晰度值得承担迁移成本时才改名；否则保留名称，并收紧文档说明。

## 术语

- Root module：顶层 `github.com/liguangsheng/goost` module。
- Nested module：带有独立 `go.mod` 的子目录，由 split-module gate 检查。
- Quick gate：日常针对单个改动面的验证路径。
- Full gate：发布或高风险变更使用的验证路径，包含较重的 race、安全、漏洞或 split-module 检查。
- Optional integration：把 `goost` 桥接到 Gin 或 gRPC 等外部框架的 nested module。
- Payload logging：integration module 中可选的 request 或 response body logging。它不是脱敏框架，必须保持显式启用。
- Retry budget：由 attempts、backoff、delay 和 context deadline 共同约束的 retry 行为预算。
- Release boundary：tag 发布前需要确认的公开文档、迁移说明、changelog、依赖图和验证命令。

## 弃用

弃用 API 使用 Go doc 的 `Deprecated:` 标记，并在移除前继续保留测试。每个弃用都需要：

- 替代方案，或明确说明没有计划替代方案；
- `MIGRATION.md` 和 `MIGRATION.zh.md` 中的迁移说明；
- 中英文 changelog 条目；
- 与当时兼容性承诺匹配的移除窗口。

pre-1.0 release 可以移除低价值 API，但移除仍然需要迁移说明和 release 文档。
