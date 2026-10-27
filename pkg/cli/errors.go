package cli

import (
	"errors"
	"fmt"
	"io"
)

// Standard sentinel errors shared across GNU coreutils implementations.
var (
	// ErrExitFailure indicates execution failed with a non-zero exit status (1).
	ErrExitFailure = errors.New("exit status 1")

	// ErrIsDirectory indicates a directory was supplied where a regular file is required.
	ErrIsDirectory = errors.New("Is a directory")

	// ErrSameInputOutput indicates an input file refers to the same file as standard output.
	ErrSameInputOutput = errors.New("input file is also the output")

	// ErrZeroLengthFileName indicates an empty filename was encountered (e.g. in --files0-from).
	ErrZeroLengthFileName = errors.New("invalid zero-length file name")

	// ErrMissingFileOperand indicates a command requiring at least one file operand received none.
	ErrMissingFileOperand = errors.New("missing file operand")
)

// ErrInvalidOption formats the GNU diagnostic for an unknown short option (-x).
func ErrInvalidOption(cmdName string, opt rune) error {
	return fmt.Errorf("%s: invalid option -- '%c'", cmdName, opt)
}

// ErrUnrecognizedOption formats the GNU diagnostic for an unknown long option (--foo).
func ErrUnrecognizedOption(cmdName, name string) error {
	return fmt.Errorf("%s: unrecognized option '--%s'", cmdName, name)
}

// ErrOptionRequiresArgShort formats the GNU diagnostic when a short option (-x) is missing its argument.
func ErrOptionRequiresArgShort(cmdName string, opt rune) error {
	return fmt.Errorf("%s: option requires an argument -- '%c'", cmdName, opt)
}

// ErrOptionRequiresArgLong formats the GNU diagnostic when a long option (--foo) is missing its argument.
func ErrOptionRequiresArgLong(cmdName, name string) error {
	return fmt.Errorf("%s: option '--%s' requires an argument", cmdName, name)
}

// ErrOptionNoArgAllowed formats the GNU diagnostic when a boolean long option is given an '=value'.
func ErrOptionNoArgAllowed(cmdName, name string) error {
	return fmt.Errorf("%s: option '--%s' doesn't allow an argument", cmdName, name)
}

// ErrCannotOpenForReading formats the GNU diagnostic when a file cannot be opened for reading.
func ErrCannotOpenForReading(filename string, err error) error {
	return fmt.Errorf("cannot open '%s' for reading: %w", filename, err)
}

// ErrReadingFile formats the GNU diagnostic when reading an open file fails.
func ErrReadingFile(filename string, err error) error {
	return fmt.Errorf("error reading '%s': %w", filename, err)
}

// ErrWrite formats a standard write failure diagnostic.
func ErrWrite(err error) error {
	return fmt.Errorf("write error: %w", err)
}

// ErrInvalidNumberOfLines formats the GNU diagnostic for an invalid line count argument.
func ErrInvalidNumberOfLines(val string) error {
	return fmt.Errorf("invalid number of lines: '%s'", val)
}

// ErrInvalidNumberOfBytes formats the GNU diagnostic for an invalid byte count argument.
func ErrInvalidNumberOfBytes(val string) error {
	return fmt.Errorf("invalid number of bytes: '%s'", val)
}

// ErrInvalidArgument formats the GNU diagnostic when an option receives an unsupported value.
func ErrInvalidArgument(val, flagName string) error {
	return fmt.Errorf("invalid argument '%s' for '%s'", val, flagName)
}

// ErrInvalidDateFormat formats the GNU diagnostic when a date/timestamp string cannot be parsed.
func ErrInvalidDateFormat(val string) error {
	return fmt.Errorf("invalid date format '%s'", val)
}

// Errorf writes a command-prefixed diagnostic line ("<Name>: <msg>\n") to stderr.
func (c *Command) Errorf(stderr io.Writer, format string, args ...any) {
	if stderr == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(stderr, "%s: %s\n", c.Name, msg)
}

// FileError writes a command- and file-prefixed diagnostic line ("<Name>: <filename>: <err>\n") to stderr.
func (c *Command) FileError(stderr io.Writer, filename string, err error) {
	if stderr == nil {
		return
	}
	fmt.Fprintf(stderr, "%s: %s: %v\n", c.Name, filename, err)
}

// TryHelpError writes a command-prefixed error followed by the standard
// "Try '<Name> --help' for more information." hint to stderr and returns ErrExitFailure.
func (c *Command) TryHelpError(stderr io.Writer, err error) error {
	if stderr != nil {
		fmt.Fprintf(stderr, "%s: %v\nTry '%s --help' for more information.\n", c.Name, err, c.Name)
	}
	return ErrExitFailure
}

