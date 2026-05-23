# 安全策略

`goost` 倾向于提供安全默认值，但它仍然是库。应用自身仍需负责认证、授权、secret 处理和数据分级。

## 日志

- `httpx` 请求摘要会刻意省略 query string 和 body。
- `httpx` retry callbacks 只暴露脱敏后的请求元数据：method、scheme、host 和 path。
- `zapctx` 和 `slogctx` 通过 context 携带 logger 和字段；它们不会自动脱敏应用传入的字段。
- Payload logging middleware 可以在配置上限内记录 request 和 response body。如果 payload 可能包含 secrets、credentials、tokens 或个人数据，请使用 `WithMaxBody(0)`、skipper functions 和 sampling。

## 文件

`rotatingwriter` 会用较严格的默认权限创建新的日志目录和文件。需要更宽访问权限的应用应提前创建路径，或在 writer 外部显式调整权限。

## 报告

私密漏洞报告请先直接联系维护者，再决定是否打开公开 issue。

