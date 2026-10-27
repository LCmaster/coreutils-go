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

// EnableAutoShortHelp enables treating -h as an automatic help option, alongside --help.
func (c *Command) EnableAutoShortHelp() {
	c.Config.AutoShortHelp = true
}

// DisableAutoShortHelp disables treating -h as an automatic help option, while preserving --help.
func (c *Command) DisableAutoShortHelp() {
	c.Config.AutoShortHelp = false
}

// EnableAutoShortVersion enables treating -v as an automatic version option, alongside --version.
func (c *Command) EnableAutoShortVersion() {
	c.Config.AutoShortVersion = true
}

// DisableAutoShortVersion disables treating -v as an automatic version option.
func (c *Command) DisableAutoShortVersion() {
	c.Config.AutoShortVersion = false
}

// EnableSoleArgumentHelpVersion configures the command parser so that --help and --version
// are only recognized when passed as the sole command-line argument.
func (c *Command) EnableSoleArgumentHelpVersion() {
	c.Config.SoleArgumentHelpVersion = true
}

// SetStopOnUnknown configures the command parser to stop option processing upon
// encountering the first non-option argument or unknown flag (used by utilities like echo).
func (c *Command) SetStopOnUnknown(stop bool) {
	c.Config.StopOnUnknown = stop
}

// EnableNumericShorthand enables parsing of legacy numeric options (e.g. -5, -20).
func (c *Command) EnableNumericShorthand(errFormat string, handler func(num string) error) {
	c.Config.NumericShorthand = handler
	if errFormat != "" {
		c.Config.NumericErrFormat = errFormat
	}
}

// SetDefaultOperands sets default operands to use when no positional arguments are given.
func (c *Command) SetDefaultOperands(operands ...string) {
	c.Config.DefaultOperands = operands
}
