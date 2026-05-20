package caseconv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ToLowerSnake(t *testing.T) {
	assert.Equal(t, "hello_world", ToLowerSnake("HelloWorld"))
	assert.Equal(t, "hello_world", ToLowerSnake("hello-world"))
	assert.Equal(t, "hello_world", ToLowerSnake("HELLO_WORLD"))
}

func Test_ToUpperCamel(t *testing.T) {
	assert.Equal(t, "HelloWorld", ToUpperCamel("hello_world"))
	assert.Equal(t, "HelloWorld", ToUpperCamel("hello-world"))
	assert.Equal(t, "HelloWorld", ToUpperCamel("HelloWorld"))
}

func Test_ToLowerCamel(t *testing.T) {
	assert.Equal(t, "helloWorld", ToLowerCamel("hello_world"))
	assert.Equal(t, "helloWorld", ToLowerCamel("hello-world"))
}

func Test_ToLowerKebab(t *testing.T) {
	assert.Equal(t, "hello-world", ToLowerKebab("HelloWorld"))
	assert.Equal(t, "hello-world", ToLowerKebab("hello_world"))
}
