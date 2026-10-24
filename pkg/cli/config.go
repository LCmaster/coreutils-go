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

// EchoConfig returns configuration tailored for the echo utility:
//   - StopOnUnknown: stops option parsing on first non-option or unknown flag.
//   - AutoShortVersion: treats -v as a version shorthand at index 0.
//   - Double dash (--) is preserved as an operand.
func EchoConfig() Config {
	cfg := DefaultConfig()
	cfg.StopOnUnknown = true
	cfg.AutoShortVersion = true
	return cfg
}

// TrueConfig returns configuration for tools like true and false:
//   - All arguments and unknown flags are ignored / treated as operands.
//   - --help, -h, --version, -v are only recognized when they are the sole argument.
func TrueConfig() Config {
	cfg := DefaultConfig()
	cfg.AutoShortVersion = true
	cfg.SoleArgumentHelpVersion = true
	return cfg
}

