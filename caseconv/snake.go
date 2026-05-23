package caseconv

import "strings"

// SnakeSplit splits s on underscores.
func SnakeSplit(s string) []string { return strings.Split(s, "_") }

// UpperSnakeJoin joins parts with underscores in upper case.
func UpperSnakeJoin(parts []string) string {
	return SimpleJoin(parts, "_", strings.ToUpper)
}

// LowerSnakeJoin joins parts with underscores in lower case.
func LowerSnakeJoin(parts []string) string {
	return SimpleJoin(parts, "_", strings.ToLower)
}

// TitleSnakeJoin joins parts with underscores, title-casing each word.
func TitleSnakeJoin(parts []string) string {
	return SimpleJoin(parts, "_", titleFirst)
}
