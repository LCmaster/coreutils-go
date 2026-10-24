package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/LCmaster/coreutils-go/pkg/ioutils"
	"github.com/LCmaster/coreutils-go/pkg/version"
)

// ErrExitFailure indicates execution failed with a non-zero exit status.
var ErrExitFailure = errors.New("exit status 1")

// RunFunc defines the execution logic for a command.
type RunFunc func(cmd *Command, stdin io.Reader, stdout, stderr io.Writer) error

// ErrorHandlerFunc defines custom error handling when parsing fails.
type ErrorHandlerFunc func(cmd *Command, stderr io.Writer, err error) int

// Options holds primitive option values managed directly by a Command.
type Options struct {
	Bools   map[string]bool
	Strings map[string]string
	Ints    map[string]int
}

// NewOptions creates an initialized Options container.
func NewOptions() *Options {
	return &Options{
		Bools:   make(map[string]bool),
		Strings: make(map[string]string),
		Ints:    make(map[string]int),
	}
}

// Bool returns the boolean value of an option by name, or false if not present.
func (o *Options) Bool(name string) bool {
	if o == nil || o.Bools == nil {
		return false
	}
	return o.Bools[name]
}

// String returns the string value of an option by name, or empty string if not present.
func (o *Options) String(name string) string {
	if o == nil || o.Strings == nil {
		return ""
	}
	return o.Strings[name]
}

// Int returns the integer value of an option by name, or 0 if not present.
func (o *Options) Int(name string) int {
	if o == nil || o.Ints == nil {
		return 0
	}
	return o.Ints[name]
}

// Command is the primary entrypoint and aggregate root for defining CLI tools.
// It coordinates command metadata (CmdHelper), flag specifications (FlagSet),
// behavioral options (Config), and delegates argument token parsing to Parser.
//
// Command.Name is the single source of truth for:
//  1. Error diagnostics: formatted as "<Name>: invalid option -- 'x'"
//  2. Help manual: default usage synopses start with "Usage: <Name> [OPTION]..."
//  3. Version output: formatted as "<Name> (coreutils-go) <version>"
type Command struct {
	Name       string // Single source of truth for errors, help, and version
	Helper     CmdHelper
	Flags      *FlagSet
	Config     Config
	Options    *Options
	Run        RunFunc
	OnError    ErrorHandlerFunc
	versionStr string
	parser     *Parser

	args      []string
	stdin     io.Reader
	stdout    io.Writer
	stderr    io.Writer
	LastError error
}

// NewCommand creates a new Command initialized with the given name and optional usage synopses.
func NewCommand(name string, usage ...string) *Command {
	cfg := DefaultConfig()
	if name == "true" || name == "false" {
		cfg = TrueConfig()
	}
	return NewCommandWithConfig(name, cfg, usage...)
}

// NewCommandWithConfig creates a new Command with an explicit configuration.
func NewCommandWithConfig(name string, cfg Config, usage ...string) *Command {
	if len(usage) == 0 && name != "" {
		usage = []string{name + " [OPTION]..."}
	}
	return &Command{
		Name:    name,
		Helper:  CmdHelper{Name: name, Usage: usage},
		Flags:   NewFlagSet(),
		Config:  cfg,
		Options: NewOptions(),
	}
}

// SetUsage sets or overrides the usage synopsis lines for the command.
func (c *Command) SetUsage(usage ...string) {
	c.Helper.Usage = usage
	for _, u := range usage {
		if strings.Contains(u, "[ignored command line arguments]") {
			c.Config.SoleArgumentHelpVersion = true
		}
	}
}

// SetDescription sets the command description for the help manual.
func (c *Command) SetDescription(desc string) {
	c.Helper.Description = desc
}

// SetNotes sets extra syntax details or format descriptions for the help manual.
func (c *Command) SetNotes(notes string) {
	c.Helper.Notes = notes
}

// AddExample adds an example usage line to the help manual.
func (c *Command) AddExample(example string) {
	c.Helper.Examples = append(c.Helper.Examples, example)
}

// AddHelpOption adds a manual / synthetic option to the help manual
// (e.g. for legacy options or options not handled by FlagSet).
func (c *Command) AddHelpOption(opt HelpOption) {
	c.Helper.ExtraFlags = append(c.Helper.ExtraFlags, opt)
}

// Version returns the standardized version string for this command,
// e.g. "<Name> (coreutils-go) 0.1.0".
func (c *Command) Version() string {
	if c.versionStr != "" {
		return c.versionStr
	}
	return version.Format(c.Name)
}

// SetVersion overrides the default version string for this command.
func (c *Command) SetVersion(v string) {
	c.versionStr = v
}

// Help formats the complete GNU help manual using registered flags and documentation metadata.
func (c *Command) Help() string {
	options := c.Flags.ToHelpOptions()

	// If auto help/version are enabled, ensure standard help/version options appear
	// if not already registered manually.
	hasHelp := false
	hasVersion := false
	for _, opt := range options {
		if strings.Contains(opt.Flags, "--help") {
			hasHelp = true
		}
		if strings.Contains(opt.Flags, "--version") {
			hasVersion = true
		}
	}
	for _, opt := range c.Helper.ExtraFlags {
		if strings.Contains(opt.Flags, "--help") {
			hasHelp = true
		}
		if strings.Contains(opt.Flags, "--version") {
			hasVersion = true
		}
	}

	var autoOpts []HelpOption
	if c.Config.AutoHelp && !hasHelp {
		short := rune(0)
		if c.Config.AutoShortHelp {
			short = 'h'
		}
		autoOpts = append(autoOpts, Opt(short, "help", "", "display this help and exit"))
	}
	if c.Config.AutoVersion && !hasVersion {
		short := rune(0)
		if c.Config.AutoShortVersion {
			short = 'v'
		}
		autoOpts = append(autoOpts, Opt(short, "version", "", "output version information and exit"))
	}

	allOptions := append(options, autoOpts...)
	return c.Helper.Format(allOptions)
}

// PrintHelp writes the formatted help manual to w.
func (c *Command) PrintHelp(w io.Writer) error {
	_, err := fmt.Fprint(w, c.Help())
	return err
}

// PrintVersion writes the formatted version string followed by a newline to w.
func (c *Command) PrintVersion(w io.Writer) error {
	_, err := fmt.Fprintln(w, c.Version())
	return err
}

// HandleAction outputs help or version to w if the parsed action is ActionHelp
// or ActionVersion. It returns handled=true if help or version was handled,
// along with any I/O error encountered during writing.
func (c *Command) HandleAction(w io.Writer) (bool, error) {
	switch c.Action() {
	case ActionHelp:
		return true, c.PrintHelp(w)
	case ActionVersion:
		return true, c.PrintVersion(w)
	default:
		return false, nil
	}
}

func (c *Command) ensureOptions() {
	if c.Options == nil {
		c.Options = NewOptions()
	}
}

// Bool returns the boolean value of the option with the given name.
func (c *Command) Bool(name string) bool {
	return c.Options.Bool(name)
}

// String returns the string value of the option with the given name.
func (c *Command) String(name string) string {
	return c.Options.String(name)
}

// Int returns the integer value of the option with the given name.
func (c *Command) Int(name string) int {
	return c.Options.Int(name)
}

// HasBool reports whether the boolean option has been explicitly recorded.
func (c *Command) HasBool(name string) bool {
	if c.Options == nil || c.Options.Bools == nil {
		return false
	}
	_, ok := c.Options.Bools[name]
	return ok
}

// HasString reports whether the string option has been explicitly recorded.
func (c *Command) HasString(name string) bool {
	if c.Options == nil || c.Options.Strings == nil {
		return false
	}
	_, ok := c.Options.Strings[name]
	return ok
}

// HasInt reports whether the integer option has been explicitly recorded.
func (c *Command) HasInt(name string) bool {
	if c.Options == nil || c.Options.Ints == nil {
		return false
	}
	_, ok := c.Options.Ints[name]
	return ok
}

// SetBool sets the boolean value of an option by name.
func (c *Command) SetBool(name string, val bool) {
	c.ensureOptions()
	c.Options.Bools[name] = val
}

// SetString sets the string value of an option by name.
func (c *Command) SetString(name string, val string) {
	c.ensureOptions()
	c.Options.Strings[name] = val
}

// SetInt sets the integer value of an option by name.
func (c *Command) SetInt(name string, val int) {
	c.ensureOptions()
	c.Options.Ints[name] = val
}

func (c *Command) recordBool(short rune, long string, val bool) {
	c.ensureOptions()
	if long != "" {
		c.Options.Bools[long] = val
	}
	if short != 0 {
		c.Options.Bools[string(short)] = val
	}
}

func (c *Command) recordString(short rune, long string, val string) {
	c.ensureOptions()
	if long != "" {
		c.Options.Strings[long] = val
	}
	if short != 0 {
		c.Options.Strings[string(short)] = val
	}
}

func (c *Command) recordInt(short rune, long string, val int) {
	c.ensureOptions()
	if long != "" {
		c.Options.Ints[long] = val
	}
	if short != 0 {
		c.Options.Ints[string(short)] = val
	}
}

// AddBool registers a boolean flag managed in c.Options.
func (c *Command) AddBool(short rune, long, description string) {
	c.AddBoolDefault(short, long, description, false)
}

// AddBoolDefault registers a boolean flag with an initial default value.
func (c *Command) AddBoolDefault(short rune, long, description string, defaultVal bool) {
	c.recordBool(short, long, defaultVal)
	c.FlagFunc(short, long, description, false, func(_ string) error {
		c.recordBool(short, long, true)
		return nil
	})
}

// AddString registers a string flag requiring an argument, managed in c.Options.
func (c *Command) AddString(short rune, long, description string) {
	c.AddStringDefault(short, long, "ARG", description, "")
}

// AddStringDefault registers a string flag requiring an argument with a default value.
func (c *Command) AddStringDefault(short rune, long, argName, description, defaultVal string) {
	c.recordString(short, long, defaultVal)
	if argName == "" {
		argName = "ARG"
	}
	c.FlagArgFunc(short, long, argName, description, func(val string) error {
		c.recordString(short, long, val)
		return nil
	})
}

// AddInt registers an integer flag requiring an argument, managed in c.Options.
func (c *Command) AddInt(short rune, long, description string) {
	c.AddIntDefault(short, long, "NUM", description, 0)
}

// AddIntDefault registers an integer flag requiring an argument with a default value.
func (c *Command) AddIntDefault(short rune, long, argName, description string, defaultVal int) {
	c.recordInt(short, long, defaultVal)
	if argName == "" {
		argName = "NUM"
	}
	c.FlagArgFunc(short, long, argName, description, func(val string) error {
		n, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid integer argument: %s", val)
		}
		c.recordInt(short, long, n)
		return nil
	})
}

// BoolVar registers a boolean flag bound to an external pointer and tracked in c.Options.
func (c *Command) BoolVar(ptr *bool, short rune, long, description string) {
	if ptr != nil {
		c.recordBool(short, long, *ptr)
	}
	c.FlagFunc(short, long, description, false, func(_ string) error {
		if ptr != nil {
			*ptr = true
		}
		c.recordBool(short, long, true)
		return nil
	})
}

// StringVar registers a string option bound to an external pointer and tracked in c.Options.
func (c *Command) StringVar(ptr *string, short rune, long, description string) {
	if ptr != nil {
		c.recordString(short, long, *ptr)
	}
	c.FlagArgFunc(short, long, "ARG", description, func(val string) error {
		if ptr != nil {
			*ptr = val
		}
		c.recordString(short, long, val)
		return nil
	})
}

// IntVar registers an integer option bound to an external pointer and tracked in c.Options.
func (c *Command) IntVar(ptr *int, short rune, long, description string) {
	if ptr != nil {
		c.recordInt(short, long, *ptr)
	}
	c.FlagArgFunc(short, long, "NUM", description, func(val string) error {
		n, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid integer argument: %s", val)
		}
		if ptr != nil {
			*ptr = n
		}
		c.recordInt(short, long, n)
		return nil
	})
}

// FlagFunc registers a custom flag with an evaluation callback.
func (c *Command) FlagFunc(short rune, long, description string, takesArg bool, handler func(val string) error) {
	argName := ""
	if takesArg {
		argName = "ARG"
	}
	c.FlagArgFunc(short, long, argName, description, handler)
}

// FlagArgFunc registers a custom flag with an argument placeholder name for help output.
func (c *Command) FlagArgFunc(short rune, long, argName, description string, handler func(val string) error) {
	c.Flags.Add(&Flag{
		Short:       short,
		Long:        long,
		ArgName:     argName,
		Description: description,
		TakesArg:    argName != "",
		Handler:     handler,
	})
}

// EnableAutoShortHelp enables treating -h as an automatic help option,
// alongside --help.
func (c *Command) EnableAutoShortHelp() {
	c.Config.AutoShortHelp = true
}

// DisableAutoShortHelp disables treating -h as an automatic help option,
// while preserving --help (e.g. for GNU cat).
func (c *Command) DisableAutoShortHelp() {
	c.Config.AutoShortHelp = false
}

// EnableAutoShortVersion enables treating -v as an automatic version option,
// alongside --version.
func (c *Command) EnableAutoShortVersion() {
	c.Config.AutoShortVersion = true
}

// DisableAutoShortVersion disables treating -v as an automatic version option.
func (c *Command) DisableAutoShortVersion() {
	c.Config.AutoShortVersion = false
}


// EnableSoleArgumentHelpVersion configures the command parser so that --help and --version
// are only recognized when passed as the sole command-line argument.
// Any other flags or arguments are treated as normal positional operands.
func (c *Command) EnableSoleArgumentHelpVersion() {
	c.Config.SoleArgumentHelpVersion = true
}

// SetStopOnUnknown configures the command parser to stop option processing upon
// encountering the first non-option argument or unknown flag (used by utilities like echo).
func (c *Command) SetStopOnUnknown(stop bool) {
	c.Config.StopOnUnknown = stop
}

// EnableNumericShorthand enables parsing of legacy numeric options (e.g. -5, -20)
// commonly supported by utilities like head and tail.
func (c *Command) EnableNumericShorthand(errFormat string, handler func(num string) error) {
	c.Config.NumericShorthand = handler
	if errFormat != "" {
		c.Config.NumericErrFormat = errFormat
	}
}

// SetDefaultOperands sets default operands to use when no positional arguments are given
// (e.g. []string{"-"} for stdin-default tools).
func (c *Command) SetDefaultOperands(operands ...string) {
	c.Config.DefaultOperands = operands
}

// Parse parses the given command-line arguments using a dedicated Parser instance.
func (c *Command) Parse(args []string) error {
	for _, u := range c.Helper.Usage {
		if strings.Contains(u, "[ignored command line arguments]") {
			c.Config.SoleArgumentHelpVersion = true
		}
	}
	c.parser = NewParser(c.Name, c.Flags, c.Config)
	return c.parser.Parse(args)
}


// Operands returns positional operands collected during Parse.
func (c *Command) Operands() []string {
	if c.parser == nil {
		return nil
	}
	return c.parser.Operands()
}

// Args is an alias for Operands matching Go standard library conventions.
func (c *Command) Args() []string {
	return c.Operands()
}

// Action returns the action determined by Parse (ActionRun, ActionHelp, or ActionVersion).
func (c *Command) Action() Action {
	if c.parser == nil {
		return ActionRun
	}
	return c.parser.Action()
}

// SetArgs explicitly sets the command-line arguments to parse instead of reading os.Args[1:].
func (c *Command) SetArgs(args []string) {
	c.args = args
}

// SetStreams sets the standard I/O streams for execution.
func (c *Command) SetStreams(stdin io.Reader, stdout, stderr io.Writer) {
	c.stdin = stdin
	c.stdout = stdout
	c.stderr = stderr
}

// SetIn sets the standard input stream.
func (c *Command) SetIn(stdin io.Reader) {
	c.stdin = stdin
}

// SetOut sets the standard output stream.
func (c *Command) SetOut(stdout io.Writer) {
	c.stdout = stdout
}

// SetErr sets the standard error stream.
func (c *Command) SetErr(stderr io.Writer) {
	c.stderr = stderr
}

// SetError records a runtime error on the command, which will be returned by Execute.
func (c *Command) SetError(err error) {
	c.LastError = err
}

// Execute parses command-line arguments, handles automatic help and version actions,
// and invokes Run if defined.
//
// Arguments default to os.Args[1:] (unless overridden via SetArgs or passed as an argument).
// Standard streams default to os.Stdin, os.Stdout, and os.Stderr (unless overridden via SetStreams).
//
// It returns nil on success (exit status 0).
// On failure, it returns an error allowing the caller to inspect or handle the error
// before exiting the program.
func (c *Command) Execute(argsAndStreams ...any) error {
	args := c.args
	if args == nil {
		if len(os.Args) > 1 {
			args = os.Args[1:]
		} else {
			args = []string{}
		}
	}
	stdin := c.stdin
	if stdin == nil {
		stdin = os.Stdin
	}
	stdout := c.stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := c.stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	if len(argsAndStreams) == 4 {
		if a, ok := argsAndStreams[0].([]string); ok {
			args = a
		}
		if in, ok := argsAndStreams[1].(io.Reader); ok {
			stdin = in
		}
		if out, ok := argsAndStreams[2].(io.Writer); ok {
			stdout = out
		}
		if errW, ok := argsAndStreams[3].(io.Writer); ok {
			stderr = errW
		}
	} else if len(argsAndStreams) != 0 {
		panic("cli.Command.Execute expects 0 arguments or 4 (args []string, stdin io.Reader, stdout, stderr io.Writer)")
	}

	return c.executeStreams(args, stdin, stdout, stderr)
}

// ExecuteWithStreams parses args and executes the command using the provided I/O streams,
// returning an error on failure.
func (c *Command) ExecuteWithStreams(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	return c.executeStreams(args, stdin, stdout, stderr)
}

func (c *Command) executeStreams(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if err := c.Parse(args); err != nil {
		if c.OnError != nil {
			c.OnError(c, stderr, err)
		} else if stderr != nil {
			fmt.Fprintf(stderr, "%v\n", err)
		}
		return err
	}

	if stdout != nil {
		if handled, err := c.HandleAction(stdout); handled {
			if err != nil && !ioutils.IsBrokenPipe(err) {
				return err
			}
			return nil
		}
	}

	if c.Run != nil {
		if err := c.Run(c, stdin, stdout, stderr); err != nil {
			return err
		}
		return nil
	}
	return nil
}




