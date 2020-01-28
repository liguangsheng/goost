package caseconv

import (
	"strings"
)

func SnakeSplit(s string) []string {
	return strings.Split(s, "_")
}

func UpperSnakeJoin(parts []string) string {
	return SimpleJoin(parts, "_", strings.ToUpper)
}

func LowerSnakeJoin(parts []string) string {
	return SimpleJoin(parts, "_", strings.ToLower)
}

func TitleSnakeJoin(parts []string) string {
	return SimpleJoin(parts, "_", strings.Title)
}
