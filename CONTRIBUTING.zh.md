# 贡献说明

`goost` 是一组小型工具库集合。变更需要保持包之间独立、依赖克制，并且能在本地清楚验证。
项目范围和弃用规则见 [PROJECT_POLICY.zh.md](./PROJECT_POLICY.zh.md)。
公开 API 约定见 [API_CONVENTIONS.zh.md](./API_CONVENTIONS.zh.md)。
验证策略见 [TESTING.zh.md](./TESTING.zh.md)。
长期 readiness 检查点见 [ROADMAP.zh.md](./ROADMAP.zh.md)。

## 提交变更前

- 先确认改动面：root package、nested module、文档、脚本，或发布边界。
- 可选集成依赖必须留在 nested modules 中。不要把 Gin、gRPC、benchmark-only 或 demo-only 依赖加入 root module。
- 公开行为、包列表、迁移说明或验证命令变化时，中英文文档要一起更新。
- 新增公开包或 integration module API 时，要增加或更新可编译示例。
- Pull request 应填写 [.github/pull_request_template.md](./.github/pull_request_template.md)，说明 API 影响、依赖影响、改动面和验证命令。Bug report 和 feature request 应使用 [.github/ISSUE_TEMPLATE/](./.github/ISSUE_TEMPLATE/) 中的模板。

## 验证

Root module 改动：

```sh
./scripts/check-root.sh --quick
```

单个 nested module 改动：

```sh
./scripts/check-split-modules.sh --quick --module <path>
```

发布前：

```sh
./scripts/check-release.sh
```
