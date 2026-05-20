# bytesconv

Zero-allocation conversion between `string` and `[]byte`. Built on
`unsafe.String` / `unsafe.SliceData`.

The returned slice / string aliases the caller's underlying memory and
must **not** be mutated.

```go
b := bytesconv.String2Bytes("hello") // []byte aliasing "hello"
s := bytesconv.Bytes2String(b)       // string aliasing b
```
