# Contributing

`goost` is a small utility-library collection. Changes should keep packages
independent, dependency-conscious, and easy to validate locally.
Project scope and deprecation rules live in
[PROJECT_POLICY.md](./PROJECT_POLICY.md).
Public API conventions live in [API_CONVENTIONS.md](./API_CONVENTIONS.md).
Validation strategy lives in [TESTING.md](./TESTING.md).
Long-term readiness checkpoints live in [ROADMAP.md](./ROADMAP.md).

## Before Opening a Change

- Identify the touched surface: root package, nested module, docs, scripts, or
  release boundary.
- Keep optional integration dependencies in nested modules. Do not add Gin,
  gRPC, benchmark-only, or demo-only dependencies to the root module.
- Update English and Chinese docs together when public behavior, package lists,
  migration notes, or validation commands change.
- Add or update compiled examples for new public packages or integration module
  APIs.
- Pull requests should fill out [.github/pull_request_template.md](./.github/pull_request_template.md),
  including API impact, dependency impact, touched surface, and validation
  commands. Bug reports and feature requests should use the issue templates in
  [.github/ISSUE_TEMPLATE/](./.github/ISSUE_TEMPLATE/).

## Validation

For root-module changes:

```sh
./scripts/check-root.sh --quick
```

For one nested module:

```sh
./scripts/check-split-modules.sh --quick --module <path>
```

Before publishing a release:

```sh
./scripts/check-release.sh
```
