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

func FormatEcho(args []string, opts *cli.Options) string {
	enableEscape := opts.Bool("EnableEscape")

	var b strings.Builder
	halt := false

	for i, arg := range args {
		if i > 0 {
			b.WriteByte(' ')
		}

		text := arg
		if enableEscape {
			text, halt = strutils.Unescape(arg)
		}
		b.WriteString(text)

		if halt {
			break
		}
	}

	if !opts.Bool("NoNewline") && !halt {
		b.WriteByte('\n')
	}

	return b.String()
}

func newCommand() *cli.Command {
	cmd := cli.NewCommand("echo")
	cmd.SetUsage("echo [OPTIONS]... [STRING]...")
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

	cmd.FlagFunc('n', "", "do not output the trailing newline", false, func(_ string) error {
		cmd.SetBool("NoNewline", true)
		return nil
	})
	cmd.FlagFunc('e', "", "enable interpretation of backslash escapes", false, func(_ string) error {
		cmd.SetBool("EnableEscape", true)
		return nil
	})
	cmd.FlagFunc('E', "", "disable interpretation of backslash escapes (default)", false, func(_ string) error {
		cmd.SetBool("EnableEscape", false)
		return nil
	})

	cmd.Run = func(c *cli.Command, _ io.Reader, stdout, _ io.Writer) error {
		writer := bufio.NewWriter(stdout)
		defer writer.Flush()
		if _, err := writer.WriteString(FormatEcho(c.Operands(), c.Options)); err != nil && !ioutils.IsBrokenPipe(err) {
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
