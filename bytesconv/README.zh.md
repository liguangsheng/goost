# bytesconv

在 `string` 和 `[]byte` 之间做零分配转换。基于 `unsafe.String` /
`unsafe.SliceData` 实现。

返回的切片或字符串会别名引用调用方的底层内存，**不得**修改。

```go
b := bytesconv.String2Bytes("hello") // []byte 别名引用 "hello"
s := bytesconv.Bytes2String(b)       // string 别名引用 b
```
