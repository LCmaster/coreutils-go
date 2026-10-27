package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LCmaster/coreutils-go/pkg/ioutils"
	"github.com/LCmaster/coreutils-go/pkg/version"
)

// RunFunc defines the execution logic for a command.
type RunFunc func(cmd *Command, stdin io.Reader, stdout, stderr io.Writer) error

// ErrorHandlerFunc defines custom error handling when parsing fails.
type ErrorHandlerFunc func(cmd *Command, stderr io.Writer, err error) int

// Command is the primary entrypoint and aggregate root for defining CLI tools.
type Command struct {
	Name   string
	Helper HelpConfig
	Flags  *FlagSet
	Config Config
	*Options
	Run        RunFunc
	OnError    ErrorHandlerFunc
	versionStr string
	parser     *Parser
}

// NewCommand creates a new Command initialized with the given name and optional usage synopses.
func NewCommand(name string, usage ...string) *Command {
	return NewCommandWithConfig(name, DefaultConfig(), usage...)
}

// NewCommandWithConfig creates a new Command with an explicit configuration.
func NewCommandWithConfig(name string, cfg Config, usage ...string) *Command {
	if len(usage) == 0 && name != "" {
		usage = []string{name + " [OPTION]..."}
	}
	cmd := &Command{
		Name:    name,
		Flags:   NewFlagSet(),
		Config:  cfg,
		Options: NewOptions(),
	}
	cmd.SetUsage(usage...)
	return cmd
}

// Version returns the standardized version string for this command.
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
	longNames := splitLongNames(long)
	normalizedLong := strings.Join(longNames, ", ")
	c.Flags.Add(&Flag{
		Short:       short,
		Long:        normalizedLong,
		ArgName:     argName,
		Description: description,
		TakesArg:    argName != "",
		Handler:     handler,
	})
}

// Parse resets c.Options and parses the given command-line arguments.
func (c *Command) Parse(args []string) error {
	c.Options = NewOptions()
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

// ExplicitOperands returns only the positional operands explicitly passed on the command line.
func (c *Command) ExplicitOperands() []string {
	if c.parser == nil {
		return nil
	}
	return c.parser.ExplicitOperands()
}

// UsedDefaultOperands reports whether DefaultOperands were substituted.
func (c *Command) UsedDefaultOperands() bool {
	if c.parser == nil {
		return false
	}
	return c.parser.UsedDefaultOperands()
}

// Action returns the action determined by Parse (ActionRun, ActionHelp, or ActionVersion).
func (c *Command) Action() Action {
	if c.parser == nil {
		return ActionRun
	}
	return c.parser.Action()
}

// Execute parses command-line arguments, handles automatic help and version actions,
// and invokes Run if defined.
func (c *Command) Execute(argsAndStreams ...any) error {
	var args []string
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}
	var stdin io.Reader = os.Stdin
	var stdout io.Writer = os.Stdout
	var stderr io.Writer = os.Stderr

	if len(argsAndStreams) == 4 {
		args, _ = argsAndStreams[0].([]string)
		stdin, _ = argsAndStreams[1].(io.Reader)
		stdout, _ = argsAndStreams[2].(io.Writer)
		stderr, _ = argsAndStreams[3].(io.Writer)
	} else if len(argsAndStreams) != 0 {
		panic("cli.Command.Execute expects 0 arguments or 4 (args []string, stdin io.Reader, stdout, stderr io.Writer)")
	}

	if err := c.Parse(args); err != nil {
		if c.OnError != nil {
			c.OnError(c, stderr, err)
		} else if stderr != nil {
			fmt.Fprintf(stderr, "%v\nTry '%s --help' for more information.\n", err, c.Name)
		}
		return err
	}

	if stdout != nil {
		switch c.Action() {
		case ActionHelp:
			if _, err := fmt.Fprint(stdout, c.Help()); err != nil && !ioutils.IsBrokenPipe(err) {
				return err
			}
			return nil
		case ActionVersion:
			if _, err := fmt.Fprintln(stdout, c.Version()); err != nil && !ioutils.IsBrokenPipe(err) {
				return err
			}
			return nil
		}
	}

	if c.Run != nil {
		return c.Run(c, stdin, stdout, stderr)
	}
	return nil
}
