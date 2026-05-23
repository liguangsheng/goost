# Project Policy

This document keeps long-lived project rules out of private planning notes.

## Scope

`goost` is a collection of small Go utility packages. It should not become a
web framework, configuration framework, logging framework, ORM, dependency
injection container, or application runtime.

Accepted additions should be reusable, dependency-conscious, and small enough
to explain without a framework manual. Heavy integrations belong in nested
modules. Demo-only and benchmark-only dependencies must stay outside the root
module.

## Addition Criteria

New packages or exported APIs should pass all of these checks before they enter
the root module:

- the purpose can be explained in one sentence without naming one application,
- at least two plausible users or packages would benefit from the API,
- the standard library or an existing `goost` package does not already cover the
  need well enough,
- the API can be documented with a small README section and a compiled example,
- the dependency impact is acceptable for the root module.

If an addition needs a framework, service SDK, heavy benchmark tool, or demo-only
dependency, keep it in a nested module, benchmark module, or example module
instead of the root module.

## Naming

Package and directory names should describe the reusable boundary, not a single
call site. Short names such as `httpx`, `zapctx`, `slogctx`, `ttlmap`, and
`defaultmap` are acceptable only while their README row and package docs define
the boundary clearly.

Before v1.0, review names that are ambiguous, too narrow, or likely to conflict
with common Go terminology. Rename only when the long-term clarity is worth the
migration cost; otherwise keep the name and tighten the docs.

## Terms

- Root module: the top-level `github.com/liguangsheng/goost` module.
- Nested module: a subdirectory with its own `go.mod`, checked by the
  split-module gate.
- Quick gate: the day-to-day validation path for one change surface.
- Full gate: the release or high-risk validation path, including heavier race,
  security, vulnerability, or split-module checks.
- Optional integration: a nested module that bridges `goost` to an external
  framework such as Gin or gRPC.
- Payload logging: optional request or response body logging in integration
  modules. It is not a redaction framework and must stay explicit.
- Retry budget: the configured attempts, backoff, delay, and context deadline
  that bound retry behavior.
- Release boundary: the public docs, migration notes, changelog, dependency
  graph, and validation commands needed before tagging a release.

## Deprecation

Deprecated APIs use the Go doc `Deprecated:` marker and stay tested until they
are removed. Each deprecation needs:

- a replacement or explicit explanation that no replacement is planned,
- migration guidance in `MIGRATION.md` and `MIGRATION.zh.md`,
- a changelog entry in both languages,
- a removal window appropriate for the compatibility promise at that time.

Pre-1.0 releases may remove low-value APIs, but removals still need migration
notes and release documentation.
