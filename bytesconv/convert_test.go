package bytesconv

import "testing"

func unused(a ...interface{}) {}

func BenchmarkString2Bytes1(b *testing.B) {
	var s = "hello, world"
	for i := 0; i < b.N; i++ {
		bytes := String2Bytes(s)
		unused(bytes)
	}
}

func BenchmarkString2Bytes2(b *testing.B) {
	var s = "hello, world"
	for i := 0; i < b.N; i++ {
		bytes := []byte(s)
		unused(bytes)
	}
}

func BenchmarkBytes2String1(b *testing.B) {
	var bytes = []byte("hello, world")
	for i := 0; i < b.N; i++ {
		s := Bytes2String(bytes)
		unused(s)
	}
}

func BenchmarkBytes2String2(b *testing.B) {
	var bytes = []byte("hello, world")
	for i := 0; i < b.N; i++ {
		s := string(bytes)
		unused(s)
	}
}
