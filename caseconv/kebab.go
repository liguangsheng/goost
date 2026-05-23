package caseconv

import "strings"

// KebabSplit splits s on hyphens.
func KebabSplit(s string) []string { return strings.Split(s, "-") }

// UpperKebabJoin joins parts with hyphens in upper case.
func UpperKebabJoin(parts []string) string {
	return SimpleJoin(parts, "-", strings.ToUpper)
}

// LowerKebabJoin joins parts with hyphens in lower case.
func LowerKebabJoin(parts []string) string {
	return SimpleJoin(parts, "-", strings.ToLower)
}

// TitleKebabJoin joins parts with hyphens, title-casing each word.
func TitleKebabJoin(parts []string) string {
	return SimpleJoin(parts, "-", titleFirst)
}
