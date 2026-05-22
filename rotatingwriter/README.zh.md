# rotatingwriter

会轮转底层文件的 `io.Writer`。

## 按天轮转

```go
w, err := rotatingwriter.NewDailyRotatingWriter("logs", "2006-01-02.log", 7)
if err != nil { /* ... */ }
log.New(w, "", log.LstdFlags).Println("hello")
```

## 按大小轮转，可选 gzip

```go
w, err := rotatingwriter.NewSizeRotatingWriter("logs/app.log", 10<<20, 5, true)
if err != nil { /* ... */ }
log.New(w, "", log.LstdFlags).Println("hello")
```

轮转后的文件名为 `app.log.1`、`app.log.2` 等。当 `gzip=true` 时，
后缀变为 `app.log.1.gz`。

## 按年龄保留

两个 rotater 都支持在数量限制之外叠加 `WithMaxAge(d)`。如果备份在轮转时
满足**任一**限制超出条件，就会被删除：

```go
// Daily：保留最新 30 个带日期文件，并删除超过 90 天的文件。
r := rotatingwriter.NewDailyRotater("logs", "2006-01-02.log", 30).
    WithMaxAge(90 * 24 * time.Hour)
w := rotatingwriter.NewRotatingWriter(r)
```

Daily rotater 的年龄来自文件名中编码的日期；size rotater 的年龄来自文件 mtime。

## 自定义策略

实现 `Rotater` 即可接入任意策略：

```go
type Rotater interface {
    Writer() io.Writer
    ShouldRollover(time.Time, int) bool // n = 待写入字节数
    DoRollover(time.Time) error
}
```

`RotatingWriter.Write` 可并发安全调用。

自动创建的日志目录使用较收敛的权限（受 umask 影响前为 `0750`）；
自动创建的日志文件和 gzip 备份使用 `0600`（受 umask 影响前）。
