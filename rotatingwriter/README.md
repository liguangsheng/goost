# rotatingwriter

An `io.Writer` that rotates the file it writes to.

## Daily rotation

```go
w, err := rotatingwriter.NewDailyRotatingWriter("logs", "2006-01-02.log", 7)
if err != nil { /* ... */ }
log.New(w, "", log.LstdFlags).Println("hello")
```

## Size-based rotation with optional gzip

```go
w, err := rotatingwriter.NewSizeRotatingWriter("logs/app.log", 10<<20, 5, true)
if err != nil { /* ... */ }
log.New(w, "", log.LstdFlags).Println("hello")
```

Rotated files are named `app.log.1`, `app.log.2`, ... When `gzip=true`,
the suffix becomes `app.log.1.gz`.

## Custom strategies

Implement `Rotater` to plug in any policy:

```go
type Rotater interface {
    Writer() io.Writer
    ShouldRollover(time.Time, int) bool // n = bytes pending
    DoRollover(time.Time) error
}
```

`RotatingWriter.Write` is safe for concurrent use.
