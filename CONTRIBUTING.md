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

## Go Version Policy

The minimum supported Go version is declared in the `go` directive of each
`go.mod` file. CI runs on the latest stable Go release. The two may differ:
`go.mod` sets the floor, CI sets the ceiling.

When upgrading Go, update all `go.mod` files, the CI `GO_VERSION` environment
variable, and `scripts/install-ci-tools.sh` in the same change.

## Hotfix Process

For urgent fixes on the current release:

1. Create a branch from the released tag (e.g., `v0.4.0`).
2. Make the minimal fix.
3. Run the quick gate for affected packages, plus `go test -race` on those
   packages.
4. Bump the patch version (e.g., `v0.4.1`) in CHANGELOG.md and CHANGELOG.zh.md.
5. Tag and push.

A full release gate is recommended but not required for patch releases when the
change is small and scoped to a single package.
