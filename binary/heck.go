package bin

import (
	"strings"
	"unicode"
)

// Ported from https://github.com/withoutboats/heck
// https://github.com/withoutboats/heck/blob/master/LICENSE-APACHE
// https://github.com/withoutboats/heck/blob/master/LICENSE-MIT

// ToPascalCase converts a string to upper camel case.
func ToPascalCase(s string) string {
	return transform(
		s,
		capitalize,
		func(*strings.Builder) {},
	)
}

// ToRustSnakeCase converts the given string to a snake_case string.
// Ported from https://github.com/withoutboats/heck/blob/c501fc95db91ce20eaef248a511caec7142208b4/src/lib.rs#L75 as used by Anchor.
func ToRustSnakeCase(s string) string {
	return transform(
		s,
		func(w string, b *strings.Builder) { b.WriteString(strings.ToLower(w)) },
		func(b *strings.Builder) { b.WriteRune('_') },
	)
}

// ToSnakeForSighash is the Anchor-sighash-ready alias for ToRustSnakeCase.
func ToSnakeForSighash(s string) string {
	return ToRustSnakeCase(s)
}

// transform walks s token-by-token, emitting a boundary callback between
// tokens and feeding each token to withWord. The word-boundary rules match
// Rust's heck crate: underscores are separators, and case changes inside a
// word (e.g. "camelCase" -> "camel" + "Case", "HTTPServer" -> "HTTP" +
// "Server") create boundaries.
func transform(
	s string,
	withWord func(string, *strings.Builder),
	boundary func(*strings.Builder),
) string {
	var builder strings.Builder
	firstWord := true

	for _, word := range splitIntoWords(s) {
		runes := []rune(word)
		init := 0
		mode := wordModeBoundary

		for i := 0; i < len(runes); i++ {
			c := runes[i]

			// Skip leading underscores within a token.
			if c == '_' {
				if init == i {
					init++
				}
				continue
			}

			if i+1 < len(runes) {
				next := runes[i+1]

				// The mode including the current character, assuming the
				// current character does not result in a word boundary.
				nextMode := mode
				switch {
				case unicode.IsLower(c):
					nextMode = wordModeLowercase
				case unicode.IsUpper(c):
					nextMode = wordModeUppercase
				}

				// Word boundary after if next is underscore, or current is
				// not uppercase and next is uppercase.
				if next == '_' || (nextMode == wordModeLowercase && unicode.IsUpper(next)) {
					if !firstWord {
						boundary(&builder)
					}
					withWord(string(runes[init:i+1]), &builder)
					firstWord = false
					init = i + 1
					mode = wordModeBoundary
					continue
				}

				// Word boundary before if current and previous are uppercase
				// and next is lowercase (XMLHttp -> XML + Http).
				if mode == wordModeUppercase && unicode.IsUpper(c) && unicode.IsLower(next) {
					if !firstWord {
						boundary(&builder)
					} else {
						firstWord = false
					}
					withWord(string(runes[init:i]), &builder)
					init = i
					mode = wordModeBoundary
					continue
				}

				// Otherwise no word boundary, just update the mode.
				mode = nextMode
				continue
			}

			// Last rune of the token: flush trailing characters as a word.
			if !firstWord {
				boundary(&builder)
			} else {
				firstWord = false
			}
			withWord(string(runes[init:]), &builder)
		}
	}

	return builder.String()
}

func capitalize(s string, b *strings.Builder) {
	if s == "" {
		return
	}
	runes := []rune(s)
	b.WriteString(strings.ToUpper(string(runes[0])))
	if len(runes) > 1 {
		lowercase(string(runes[1:]), b)
	}
}

func lowercase(s string, b *strings.Builder) {
	runes := []rune(s)
	for i, c := range runes {
		// Final sigma special-case per the heck crate.
		if c == 'Σ' && i == len(runes)-1 {
			b.WriteString("ς")
			continue
		}
		b.WriteString(strings.ToLower(string(c)))
	}
}

func splitIntoWords(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
}

type wordMode int

const (
	// wordModeBoundary: no lowercase or uppercase characters yet in the
	// current word.
	wordModeBoundary wordMode = iota
	// wordModeLowercase: the previous cased character in the current word is
	// lowercase.
	wordModeLowercase
	// wordModeUppercase: the previous cased character in the current word is
	// uppercase.
	wordModeUppercase
)
