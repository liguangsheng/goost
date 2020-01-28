package caseconv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CamelSplit(t *testing.T) {
	ast := assert.New(t)
	ast.Equal([]string{"Hello", "HTTP", "url", "ID"}, CamelSplit("HelloHTTPurlID"))
	ast.Equal([]string{"Hello", "HTTP", "World"}, CamelSplit("HelloHTTPWorld"))
}

func Test_CamelJoin(t *testing.T) {
	ast := assert.New(t)
	ast.Equal("AaaHTTPurlID", CamelJoin([]string{"aaa", "http", "url", "id"}, true))
	ast.Equal("aaaHTTPurlID", CamelJoin([]string{"aaa", "http", "url", "id"}, false))
}

func Test_UpperCamelJoin(t *testing.T) {
	ast := assert.New(t)
	ast.Equal("AaaHTTPurlID", UpperCamelJoin([]string{"aaa", "http", "url", "id"}))
}

func Test_LowerCamelJoin(t *testing.T) {
	ast := assert.New(t)
	ast.Equal("aaaHTTPurlID", LowerCamelJoin([]string{"aaa", "http", "url", "id"}))
}
