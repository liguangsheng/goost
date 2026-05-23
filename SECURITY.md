# Security Policy

`goost` favors safe defaults, but it is still a library. Applications remain
responsible for their own authentication, authorization, secret handling, and
data classification.

## Logging

- `httpx` request summaries intentionally omit query strings and bodies.
- `httpx` retry callbacks expose sanitized request metadata: method, scheme,
  host, and path.
- `zapctx` and `slogctx` carry loggers and fields through context; they do not
  automatically redact application-provided fields.
- Payload logging middleware can record request and response bodies up to the
  configured limit. Use `WithMaxBody(0)`, skipper functions, and sampling when
  payloads may contain secrets, credentials, tokens, or personal data.

## Files

`rotatingwriter` creates new log directories and files with restrictive default
permissions. Applications that need broader access should pre-create paths or
adjust permissions explicitly outside the writer.

## Reporting

For private vulnerability reports, contact the maintainer directly before
opening a public issue.

