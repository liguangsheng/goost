// Package bytesconv provides allocation-free conversions between string and
// []byte using unsafe. The returned slice or string aliases the input's
// underlying memory; mutating the slice obtained from String2Bytes or
// modifying the byte slice passed to Bytes2String results in undefined
// behavior.
package bytesconv

import "unsafe"

// String2Bytes returns a byte slice that aliases s's bytes. The returned
// slice MUST NOT be modified.
func String2Bytes(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// Bytes2String returns a string that aliases b's bytes. b MUST NOT be
// modified for the lifetime of the returned string.
func Bytes2String(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(b), len(b))
}
