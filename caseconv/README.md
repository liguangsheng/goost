# caseconv

Split and join identifiers across casing styles.

```go
caseconv.CamelSplit("HelloHTTPWorld")               // ["Hello", "HTTP", "World"]
caseconv.UpperCamelJoin([]string{"i","love","you"}) // "ILoveYou"
caseconv.LowerCamelJoin([]string{"i","love","you"}) // "iLoveYou"
caseconv.UpperSnakeJoin([]string{"i","love","you"}) // "I_LOVE_YOU"
caseconv.LowerSnakeJoin([]string{"i","love","you"}) // "i_love_you"
caseconv.UpperKebabJoin([]string{"i","love","you"}) // "I-LOVE-YOU"
caseconv.TitleSnakeJoin([]string{"i","love","you"}) // "I_Love_You"
```

Acronyms are preserved in camel-case output. Register additional ones at
startup; the default set includes `HTTP`, `ID`, `WWW`, `URL`, `DAO`, `XML`:

```go
caseconv.RegisterAcronym("ACL")
```
