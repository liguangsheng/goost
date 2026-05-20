package bytesconv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_String2Bytes(t *testing.T) {
	s := "hello, world"
	b := String2Bytes(s)
	assert.Equal(t, []byte(s), b)
}

func Test_Bytes2String(t *testing.T) {
	b := []byte("hello, world")
	s := Bytes2String(b)
	assert.Equal(t, string(b), s)
}

func Test_EmptyString2Bytes(t *testing.T) {
	assert.Nil(t, String2Bytes(""))
}

func Test_EmptyBytes2String(t *testing.T) {
	assert.Equal(t, "", Bytes2String(nil))
	assert.Equal(t, "", Bytes2String([]byte{}))
}

func Test_RoundTrip(t *testing.T) {
	s := "go is fun"
	assert.Equal(t, s, Bytes2String(String2Bytes(s)))
}

func unused(a ...any) {}

func BenchmarkString2Bytes1(b *testing.B) {
	s := "hello, world"
	for range b.N {
		bs := String2Bytes(s)
		unused(bs)
	}
}

func BenchmarkString2Bytes2(b *testing.B) {
	s := "hello, world"
	for range b.N {
		bs := []byte(s)
		unused(bs)
	}
}

func BenchmarkBytes2String1(b *testing.B) {
	bs := []byte("hello, world")
	for range b.N {
		s := Bytes2String(bs)
		unused(s)
	}
}

func BenchmarkBytes2String2(b *testing.B) {
	bs := []byte("hello, world")
	for range b.N {
		s := string(bs)
		unused(s)
	}
}
