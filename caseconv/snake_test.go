package caseconv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SnakeSplit(t *testing.T) {
	ast := assert.New(t)
	ast.Equal([]string{"aaa", "bbb", "ccc"}, SnakeSplit("aaa_bbb_ccc"))
}

func Test_UpperSnakeJoin(t *testing.T) {
	ast := assert.New(t)
	ast.Equal("AAA_BBB_CCC", UpperSnakeJoin([]string{"aaa", "bbb", "ccc"}))
}

func Test_LowerSnakeJoin(t *testing.T) {
	ast := assert.New(t)
	ast.Equal("aaa_bbb_ccc", LowerSnakeJoin([]string{"aaa", "bbb", "ccc"}))
}

func Test_TitleSnakeJoin(t *testing.T) {
	ast := assert.New(t)
	ast.Equal("Aaa_Bbb_Ccc", TitleSnakeJoin([]string{"aaa", "bbb", "ccc"}))
}
