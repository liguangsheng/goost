# rotating_writer

An `io.Writer` that rotates the file it writes to.

```go
w, err := rotating_writer.NewDailyRotatingWriter("logs", "2006-01-02.log", 7)
if err != nil { /* ... */ }

logger := log.New(w, "", log.LstdFlags)
logger.Println("hello")
```

Provide a custom `Rotater` to implement other policies (size-based, hourly,
etc.). `RotatingWriter.Write` is safe for concurrent use.
