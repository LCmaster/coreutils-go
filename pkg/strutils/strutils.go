// Package strutils provides low-level string and byte-slice manipulation routines
// adhering to POSIX and GNU core utilities conventions.
package strutils

// FormatNonPrinting converts control characters and non-ASCII bytes into visible
// representations using POSIX/GNU Caret ('^') and Meta ("M-") notation.
//
// Formatting rules:
//   - Tab ('\t') and Newline ('\n') are preserved as literal characters.
//   - Control characters with byte values in [0, 31] are formatted as '^' followed
//     by the byte shifted by +64 (e.g. 0x00 -> "^@", 0x01 -> "^A", 0x07 -> "^G").
//   - Delete character 127 (0x7F) is formatted as "^?".
//   - High-order control bytes in [128, 159] are formatted as "M-^" followed by
//     the byte shifted by -128 + 64 (e.g. 0x80 -> "M-^@").
//   - High-order bytes in [160, 254] are formatted as "M-" followed by the byte
//     shifted by -128 (e.g. 0xA0 -> "M- ").
//   - Byte 255 (0xFF) is formatted as "M-^?".
//
// This matches the behavior of GNU `cat -v` and `ls -b` (`--show-control-chars`).
func FormatNonPrinting(line []byte) []byte {
	result := make([]byte, 0, len(line))
	for _, b := range line {
		switch {
		case b == '\n' || b == '\t':
			result = append(result, b)
		case b < 32:
			result = append(result, '^', b+64)
		case b < 127:
			result = append(result, b)
		case b == 127:
			result = append(result, '^', '?')
		case b < 160:
			result = append(result, 'M', '-', '^', b-128+64)
		case b < 255:
			result = append(result, 'M', '-', b-128)
		default: // b == 255
			result = append(result, 'M', '-', '^', '?')
		}
	}
	return result
}

// IsEmptyLine reports whether line contains no content other than an optional line
// termination sequence.
//
// It returns true if line is empty (len == 0), consists solely of a Unix newline ("\n"),
// or consists solely of a Windows/DOS carriage-return and newline ("\r\n").
// Any line containing spaces, tabs, or printable characters returns false.
//
// This check performs zero memory allocations and is used by formatting pipelines
// such as `cat -s` (squeeze-blank) and `cat -b` (number-nonblank).
func IsEmptyLine(line []byte) bool {
	return len(line) == 0 || (len(line) == 1 && line[0] == '\n') || (len(line) == 2 && line[0] == '\r' && line[1] == '\n')
}
