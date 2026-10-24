package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LCmaster/coreutils-go/pkg/cli"
	"github.com/LCmaster/coreutils-go/pkg/ioutils"
	"github.com/LCmaster/coreutils-go/pkg/strutils"
	"github.com/LCmaster/coreutils-go/pkg/version"
)

type CLIAction int

const (
	ActionRun CLIAction = iota
	ActionHelp
	ActionVersion
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

func isEchoFlagGroup(arg string) bool {
	if len(arg) < 2 || arg[0] != '-' {
		return false
	}
	for _, ch := range arg[1:] {
		if ch != 'n' && ch != 'e' && ch != 'E' {
			return false
		}
	}
	return true
}

func parseArgs(rawArgs []string) (EchoOptions, []string, CLIAction) {
	var opts EchoOptions
	i := 0

	for i < len(rawArgs) {
		arg := rawArgs[i]

		if i == 0 {
			switch arg {
			case "-h", "--help":
				return opts, nil, ActionHelp
			case "-v", "--version":
				return opts, nil, ActionVersion
			}
		}

		if !isEchoFlagGroup(arg) {
			break
		}

		for _, ch := range arg[1:] {
			switch ch {
			case 'n':
				opts.NoNewline = true
			case 'e':
				opts.EnableEscape = true
			case 'E':
				opts.EnableEscape = false
			}
		}
		i++
	}

	return opts, rawArgs[i:], ActionRun
}

func getHelpStr() string {
	return cli.FormatHelp(cli.HelpConfig{
		Usage:       []string{"echo [OPTIONS]... [STRING]..."},
		Description: "echo writes each given string to standard output, with a space between each and a newline after the last one.",
		Options: []cli.HelpOption{
			cli.Opt('n', "", "", "do not output the trailing newline"),
			cli.Opt('e', "", "", "enable interpretation of backslash escapes"),
			cli.Opt('E', "", "", "disable interpretation of backslash escapes (default)"),
			cli.Opt('h', "help", "", "display this help and exit"),
			cli.Opt('v', "version", "", "output version information and exit"),
		},
		Notes: "If -e is in effect, the following sequences are recognized:\n" +
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
			"  \\xHH\t\tbyte with hexadecimal value HH (1 to 2 digits)",
	})
}

func getVersionStr() string {
	return version.Format("echo")
}

func run(args []string, _ io.Reader, stdout, _ io.Writer) int {
	opts, stringArgs, action := parseArgs(args)

	switch action {
	case ActionVersion:
		if _, err := fmt.Fprintln(stdout, getVersionStr()); err != nil && !ioutils.IsBrokenPipe(err) {
			return 1
		}
	case ActionHelp:
		if _, err := fmt.Fprint(stdout, getHelpStr()); err != nil && !ioutils.IsBrokenPipe(err) {
			return 1
		}
	case ActionRun:
		writer := bufio.NewWriter(stdout)
		defer writer.Flush()
		if _, err := writer.WriteString(FormatEcho(stringArgs, opts)); err != nil && !ioutils.IsBrokenPipe(err) {
			return 1
		}
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
