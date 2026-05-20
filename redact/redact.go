// Package redact provides string-redaction helpers that keep enough of
// the original visible for debugging while removing sensitive content.
//
// All functions are pure and zero-dependency. For zap and slog
// integration, see ZapString and SlogString.
package redact

import "strings"

// Mask returns s with all but the first keepLeft and last keepRight
// characters replaced by '*'. If s is shorter than keepLeft+keepRight,
// every character is masked.
func Mask(s string, keepLeft, keepRight int) string {
	if keepLeft < 0 {
		keepLeft = 0
	}
	if keepRight < 0 {
		keepRight = 0
	}
	r := []rune(s)
	if len(r) <= keepLeft+keepRight {
		return strings.Repeat("*", len(r))
	}
	var b strings.Builder
	b.Grow(len(r))
	b.WriteString(string(r[:keepLeft]))
	b.WriteString(strings.Repeat("*", len(r)-keepLeft-keepRight))
	b.WriteString(string(r[len(r)-keepRight:]))
	return b.String()
}

// Email redacts the local-part of an email, keeping the first character
// and the entire domain: "alice@example.com" -> "a****@example.com".
func Email(s string) string {
	at := strings.LastIndexByte(s, '@')
	if at <= 0 {
		return Mask(s, 1, 0)
	}
	local := s[:at]
	domain := s[at:]
	return Mask(local, 1, 0) + domain
}

// Phone keeps the first 3 and last 4 digits/characters: "13800138000" ->
// "138****8000". Non-digit characters are preserved up to those windows.
func Phone(s string) string {
	return Mask(s, 3, 4)
}

// Token masks all but the first 4 and last 4 characters; intended for
// API keys, JWTs, and similar.
func Token(s string) string {
	return Mask(s, 4, 4)
}
