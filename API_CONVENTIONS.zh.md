# API 约定

这些规则用于指导 v1.0 前的新公开 API 和清理工作。

## 构造器与零值

- 拥有 goroutine、timer、锁、map、queue 或配置的类型，必须说明 zero value 是否可用。
- 如果 zero value 不适合直接使用，优先提供明确 constructor 或 builder，并在类型 doc comment 中说明。
- 如果无效配置最终会在 goroutine 中才失败，constructor 应尽早拒绝它。

## 泛型

- 泛型 API 应保持类型参数最少，并让普通调用尽量容易类型推断。
- Comparator、loader、callback 和 hook 的签名如果有不明显的所有权或并发预期，需要在文档中说明。
- 如果期望类型推断可用，示例应尽量不需要显式类型参数也能编译。

## Context、取消与生命周期

- `context.Context` 参数默认只控制当前操作；如果它还控制长生命周期对象，API 必须明确说明。
- 如果取消是公开契约的一部分，调用方应能及时返回，并收到 `ctx.Err()` 或兼容 `errors.Is` 的错误。
- 会启动 goroutine 或 timer 的 constructor / builder，必须记录对应的 `Close`、`Stop`、`Wait` 或其他释放路径。
- 除非包文档说明了更窄的契约，`Close` 和 `Stop` 应支持重复调用。
- 后台 goroutine 应通过文档中的生命周期路径停止，不能要求调用方依赖 garbage collection。

## 观测能力

- `Stats`、`Snapshot` 和类似值，对调用方来说应是不可变快照，而不是指向内部可变状态的 live view。
- 字段文档应说明每个值是当前 gauge、累计 counter、配置值还是派生值。
- 如果拥有者类型本身支持并发使用，snapshot 方法也必须能与普通操作并发调用。
- Callback 和 hook API 应说明它们是否同步执行、是否可能阻塞主流程，以及 panic 是否会被 recover。

## 错误与 panic

- Sentinel errors 使用 `Err` 前缀（例如 `ErrPoolClosed`、`ErrOpen`）。
- Sentinel error message 使用包名前缀（例如 `"pool: closed"`、
  `"circuitbreaker: open"`），方便在日志中识别来源。
- Sentinel errors 应兼容标准库 `errors.Is`。
- Wrapped errors 应通过 `fmt.Errorf("pkg: %w", err)` 保留
  `errors.Is` / `errors.As` 行为。
- 如果包会 recover panic，必须记录这个边界，并通过 error 或 callback value 暴露恢复到的 panic。
- 如果 context cancellation 是公开操作的一部分，应作为普通 error 返回。

## Hook 与 Callback

- Hook 和 callback 函数会同步运行在调用方 goroutine 或包内部 goroutine 上；每个包必须说明具体位置。
- Hook 内部的 panic 不能导致宿主操作崩溃。包应 recover hook panic，并保持宿主操作的正常结果。
- 可能阻塞进度的 hook 必须记录阻塞风险，方便调用方决定是否自行转交到 goroutine。
