package cli

import (
	"strings"
)

// splitLongNames splits a comma-separated list of long flag aliases,
// trimming whitespace and leading "--".
func splitLongNames(long string) []string {
	if long == "" {
		return nil
	}
	parts := strings.Split(long, ",")
	names := make([]string, 0, len(parts))
	for _, p := range parts {
		name := strings.TrimSpace(p)
		name = strings.TrimPrefix(name, "--")
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

// Flag represents a command-line flag specification.
type Flag struct {
	Short       rune
	Long        string
	ArgName     string // Optional argument placeholder for help (e.g. "NUM", "FILE", "[-]NUM")
	Description string
	TakesArg    bool
	Handler     func(val string) error
}

// FlagSet manages a collection of registered command-line flags and fast lookups.
type FlagSet struct {
	flags    []*Flag
	shortMap map[rune]*Flag
	longMap  map[string]*Flag
}

// NewFlagSet initializes an empty FlagSet.
func NewFlagSet() *FlagSet {
	return &FlagSet{
		shortMap: make(map[rune]*Flag),
		longMap:  make(map[string]*Flag),
	}
}

// Add registers a Flag in the FlagSet.
func (fs *FlagSet) Add(f *Flag) {
	fs.flags = append(fs.flags, f)
	if f.Short != 0 {
		fs.shortMap[f.Short] = f
	}
	if f.Long != "" {
		for _, name := range splitLongNames(f.Long) {
			fs.longMap[name] = f
		}
	}
}

// LookupShort returns the Flag registered for the given short rune, if any.
func (fs *FlagSet) LookupShort(r rune) (*Flag, bool) {
	f, ok := fs.shortMap[r]
	return f, ok
}

// LookupLong returns the Flag registered for the given long name, if any.
func (fs *FlagSet) LookupLong(name string) (*Flag, bool) {
	f, ok := fs.longMap[name]
	return f, ok
}

// Flags returns a slice of all registered flags in insertion order.
func (fs *FlagSet) Flags() []*Flag {
	return fs.flags
}

// ToHelpOptions converts registered flags to HelpOption structs for help generation.
func (fs *FlagSet) ToHelpOptions() []HelpOption {
	opts := make([]HelpOption, 0, len(fs.flags))
	for _, f := range fs.flags {
		opts = append(opts, Opt(f.Short, f.Long, f.ArgName, f.Description))
	}
	return opts
}
