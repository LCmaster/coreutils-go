// Package unescape provides utilities for interpreting escape sequences modeled after GNU coreutils.
package unescape

import (
	"strings"
)

// Unescape parses backslash escape sequences in s according to GNU echo rules.
// It returns the unescaped string and a boolean indicating whether '\c' was encountered.
// If halt is true, all subsequent output must cease and trailing newlines must be suppressed.
func Unescape(s string) (string, bool) {
	var buf strings.Builder
	buf.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
			switch s[i] {
			case '\\':
				buf.WriteByte('\\')
			case 'a':
				buf.WriteByte('\a') // alert (BEL, 7)
			case 'b':
				buf.WriteByte('\b') // backspace (8)
			case 'c':
				// produce no further output
				return buf.String(), true
			case 'e':
				buf.WriteByte(0x1B) // escape (27)
			case 'f':
				buf.WriteByte('\f') // form feed (12)
			case 'n':
				buf.WriteByte('\n') // new line (10)
			case 'r':
				buf.WriteByte('\r') // carriage return (13)
			case 't':
				buf.WriteByte('\t') // horizontal tab (9)
			case 'v':
				buf.WriteByte('\v') // vertical tab (11)
			case '0':
				// \0NNN: octal byte value with 0 to 3 octal digits following '0'
				val := 0
				count := 0
				for count < 3 && i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '7' {
					i++
					val = val*8 + int(s[i]-'0')
					count++
				}
				buf.WriteByte(byte(val & 0xFF))
			case 'x':
				// \xHH: hexadecimal byte value with 1 to 2 hex digits
				val := 0
				count := 0
				for count < 2 && i+1 < len(s) {
					c := s[i+1]
					var digit int
					if c >= '0' && c <= '9' {
						digit = int(c - '0')
					} else if c >= 'a' && c <= 'f' {
						digit = int(c - 'a' + 10)
					} else if c >= 'A' && c <= 'F' {
						digit = int(c - 'A' + 10)
					} else {
						break
					}
					i++
					val = val*16 + digit
					count++
				}
				if count == 0 {
					buf.WriteString(`\x`)
				} else {
					buf.WriteByte(byte(val & 0xFF))
				}
			default:
				// Unrecognized escape: write literal \ followed by character
				buf.WriteByte('\\')
				buf.WriteByte(s[i])
			}
		} else {
			buf.WriteByte(s[i])
		}
	}

	return buf.String(), false
}
