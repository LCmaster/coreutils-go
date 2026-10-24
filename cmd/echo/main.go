package main

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/LCmaster/coreutils-go/pkg/cli"
	"github.com/LCmaster/coreutils-go/pkg/ioutils"
	"github.com/LCmaster/coreutils-go/pkg/strutils"
)

type EchoOptions struct {
	NoNewline    bool
	EnableEscape bool
}

type ArgTransformer func(string) (output string, halt bool)

func IdentityTransformer(s string) (string, bool) {
	return s, false
}

func EscapeTransformer(s string) (string, bool) {
	return strutils.Unescape(s)
}

func FormatEcho(args []string, opts EchoOptions) string {
	transform := IdentityTransformer
	if opts.EnableEscape {
		transform = EscapeTransformer
	}

	var b strings.Builder
	halt := false

	for i, arg := range args {
		if i > 0 {
			b.WriteByte(' ')
		}

		var text string
		text, halt = transform(arg)
		b.WriteString(text)

		if halt {
			break
		}
	}

	if !opts.NoNewline && !halt {
		b.WriteByte('\n')
	}

	return b.String()
}

func newCommand() *cli.Command {
	var opts EchoOptions

	cmd := cli.NewCommand("echo")
	cmd.Helper.SetUsage("echo [OPTIONS]... [STRING]...")
	cmd.SetStopOnUnknown(true)

	cmd.EnableAutoShortVersion()
	cmd.SetDescription("echo writes each given string to standard output, with a space between each and a newline after the last one.")
	cmd.SetNotes("If -e is in effect, the following sequences are recognized:\n" +
		"  \\\\\t\tbackslash\n" +
		"  \\a\t\talert (BEL)\n" +
		"  \\b\t\tbackspace\n" +
		"  \\c\t\tproduce no further output\n" +
		"  \\e\t\tescape\n" +
		"  \\f\t\tform feed\n" +
		"  \\n\t\tnew line\n" +
		"  \\r\t\tcarriage return\n" +
		"  \\t\t\thorizontal tab\n" +
		"  \\v\t\tvertical tab\n" +
		"  \\0NNN\t\tbyte with octal value NNN (1 to 3 digits)\n" +
		"  \\xHH\t\tbyte with hexadecimal value HH (1 to 2 digits)")

	cmd.BoolVar(&opts.NoNewline, 'n', "", "do not output the trailing newline")
	cmd.FlagFunc('e', "", "enable interpretation of backslash escapes", false, func(_ string) error {
		opts.EnableEscape = true
		return nil
	})
	cmd.FlagFunc('E', "", "disable interpretation of backslash escapes (default)", false, func(_ string) error {
		opts.EnableEscape = false
		return nil
	})

	cmd.Run = func(c *cli.Command, _ io.Reader, stdout, _ io.Writer) error {
		writer := bufio.NewWriter(stdout)
		defer writer.Flush()
		if _, err := writer.WriteString(FormatEcho(c.Operands(), opts)); err != nil && !ioutils.IsBrokenPipe(err) {
			return err
		}
		return nil
	}
	return cmd
}

func main() {
	if err := newCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
