package cli

// Config encapsulates behavioral settings for CLI argument parsing.
type Config struct {
	AutoHelp                bool
	AutoShortHelp           bool
	AutoVersion             bool
	AutoShortVersion        bool
	StopOnUnknown           bool
	SoleArgumentHelpVersion bool
	DefaultOperands         []string
	NumericShorthand        func(num string) error
	NumericErrFormat        string
}

// DefaultConfig returns the standard POSIX/GNU coreutils parser configuration:
//   - Interspersed options and operands are supported.
//   - --help and -h automatically trigger ActionHelp.
//   - --version automatically triggers ActionVersion.
func DefaultConfig() Config {
	return Config{
		AutoHelp:         true,
		AutoShortHelp:    true,
		AutoVersion:      true,
		AutoShortVersion: false,
		StopOnUnknown:    false,
		NumericErrFormat: "%s: invalid number: '%s'",
	}
}
