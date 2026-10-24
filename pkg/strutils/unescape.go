package strutils

import (
	"strings"
)

// Unescape parses backslash escape sequences in string s according to GNU `echo -e` rules.
//
// Supported escape sequences:
//   - `\\`   : literal backslash
//   - `\a`   : alert (BEL, ASCII 7)
//   - `\b`   : backspace (ASCII 8)
//   - `\c`   : produce no further output and suppress any trailing newline
//   - `\e`   : escape (ESC, ASCII 27 / 0x1B)
//   - `\f`   : form feed (ASCII 12)
//   - `\n`   : new line (LF, ASCII 10)
//   - `\r`   : carriage return (CR, ASCII 13)
//   - `\t`   : horizontal tab (ASCII 9)
//   - `\v`   : vertical tab (ASCII 11)
//   - `\0NNN`: byte with octal value NNN (0 to 3 octal digits following '0')
//   - `\xHH` : byte with hexadecimal value HH (1 to 2 hexadecimal digits)
//
// Unrecognized escape sequences (e.g. `\z`) are output literally as the backslash
// followed by the character (e.g. `\z`). If a trailing backslash occurs at the end
// of the string without a following character, it is treated as a literal backslash.
//
// Return values:
//   - out : the unescaped string constructed up to the point of completion or early halt.
//   - halt: true if the `\c` escape was encountered, signaling that all subsequent
//     output processing must stop immediately and trailing newlines must be omitted.
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
