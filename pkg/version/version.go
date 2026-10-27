// Package version provides centralized version metadata and formatting
// for the coreutils-go suite.
package version

import "fmt"

// Current is the semantic version of the coreutils-go suite.
// It is declared as a variable so release builds can override it via -ldflags "-X".
var Current = "0.1.0"

// Format returns the standard version banner for a given core utility name.
// e.g. Format("cat") returns "cat (coreutils-go) 0.1.0".
func Format(utility string) string {
	return fmt.Sprintf("%s (coreutils-go) %s", utility, Current)
}
