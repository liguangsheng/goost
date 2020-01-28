package caseconv

import (
	"strings"
)

func KebabSplit(s string) []string {
	return strings.Split(s, "-")
}

func UpperKebabJoin(parts []string) string {
	return SimpleJoin(parts, "-", strings.ToUpper)
}

func LowerKebabJoin(parts []string) string {
	return SimpleJoin(parts, "-", strings.ToLower)
}

func TitleKebabJoin(parts []string) string {
	return SimpleJoin(parts, "-", strings.Title)
}
