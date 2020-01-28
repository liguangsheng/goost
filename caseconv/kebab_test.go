package caseconv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_KebabSplit(t *testing.T) {
	ast := assert.New(t)
	ast.Equal([]string{"aaa", "bbb", "ccc"}, KebabSplit("aaa-bbb-ccc"))
}

func Test_UpperKebabJoin(t *testing.T) {
	ast := assert.New(t)
	ast.Equal("AAA-BBB-CCC", UpperKebabJoin([]string{"aaa", "bbb", "ccc"}))
}

func Test_LowerKebabJoin(t *testing.T) {
	ast := assert.New(t)
	ast.Equal("aaa-bbb-ccc", LowerKebabJoin([]string{"aaa", "bbb", "ccc"}))
}

func Test_TitleKebabJoin(t *testing.T) {
	ast := assert.New(t)
	ast.Equal("Aaa-Bbb-Ccc", TitleKebabJoin([]string{"aaa", "bbb", "ccc"}))
}
