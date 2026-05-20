package caseconv

import "strings"

// split detects which case s is in and returns its constituent words.
// Heuristic: presence of '_' picks snake; '-' picks kebab; otherwise camel.
func split(s string) []string {
	switch {
	case strings.ContainsRune(s, '_'):
		return SnakeSplit(s)
	case strings.ContainsRune(s, '-'):
		return KebabSplit(s)
	default:
		return CamelSplit(s)
	}
}

// ToUpperCamel converts any cased identifier to UpperCamelCase.
func ToUpperCamel(s string) string { return UpperCamelJoin(split(s)) }

// ToLowerCamel converts any cased identifier to lowerCamelCase.
func ToLowerCamel(s string) string { return LowerCamelJoin(split(s)) }

// ToPascal is an alias of ToUpperCamel.
func ToPascal(s string) string { return ToUpperCamel(s) }

// ToLowerSnake converts any cased identifier to lower_snake_case.
func ToLowerSnake(s string) string { return LowerSnakeJoin(split(s)) }

// ToUpperSnake converts any cased identifier to UPPER_SNAKE_CASE.
func ToUpperSnake(s string) string { return UpperSnakeJoin(split(s)) }

// ToTitleSnake converts any cased identifier to Title_Snake_Case.
func ToTitleSnake(s string) string { return TitleSnakeJoin(split(s)) }

// ToLowerKebab converts any cased identifier to lower-kebab-case.
func ToLowerKebab(s string) string { return LowerKebabJoin(split(s)) }

// ToUpperKebab converts any cased identifier to UPPER-KEBAB-CASE.
func ToUpperKebab(s string) string { return UpperKebabJoin(split(s)) }

// ToTitleKebab converts any cased identifier to Title-Kebab-Case.
func ToTitleKebab(s string) string { return TitleKebabJoin(split(s)) }
