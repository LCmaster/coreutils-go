package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LCmaster/coreutils-go/pkg/unescape"
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
	return unescape.Unescape(s)
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

func printHelp() string {
	helpMessage := "" +
		"Usage:\n" +
		"echo [OPTIONS]... [STRING]...\n" +
		"\n" +
		"Description:\n" +
		"echo writes each given string to standard output, with a space between each and a newline after the last one.\n" +
		"  -n\t\t\tdo not output the trailing newline\n" +
		"  -e\t\t\tenable interpretation of backslash escapes\n" +
		"  -E\t\t\tdisable interpretation of backslash escapes (default)\n" +
		"  -h, --help\tdisplay this help and exit\n" +
		"  -v, --version\toutput version information and exit\n" +
		"\n" +
		"If -e is in effect, the following sequences are recognized:\n" +
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
		"  \\xHH\t\tbyte with hexadecimal value HH (1 to 2 digits)\n"

	return helpMessage
}

func printVersion(w io.Writer) {
	fmt.Fprintln(w, "echo (coreutils-go) 0.1.0")
}

func run(args []string, stdout io.Writer) int {
	opts, stringArgs, action := parseArgs(args)

	switch action {
	case ActionVersion:
		printVersion(stdout)
	case ActionHelp:
		fmt.Fprint(stdout, printHelp())
	case ActionRun:
		writer := bufio.NewWriter(stdout)
		defer writer.Flush()
		writer.WriteString(FormatEcho(stringArgs, opts))
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout))
}
