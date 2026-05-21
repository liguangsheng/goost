# caseconv

在不同命名风格之间拆分和拼接标识符。

```go
caseconv.CamelSplit("HelloHTTPWorld")               // ["Hello", "HTTP", "World"]
caseconv.UpperCamelJoin([]string{"i","love","you"}) // "ILoveYou"
caseconv.LowerCamelJoin([]string{"i","love","you"}) // "iLoveYou"
caseconv.UpperSnakeJoin([]string{"i","love","you"}) // "I_LOVE_YOU"
caseconv.LowerSnakeJoin([]string{"i","love","you"}) // "i_love_you"
caseconv.UpperKebabJoin([]string{"i","love","you"}) // "I-LOVE-YOU"
caseconv.TitleSnakeJoin([]string{"i","love","you"}) // "I_Love_You"
```

缩略词会在 camel-case 输出中保留。可以在启动时注册额外缩略词；
默认集合包含 `HTTP`、`ID`、`WWW`、`URL`、`DAO`、`XML`：

```go
caseconv.RegisterAcronym("ACL")
```
