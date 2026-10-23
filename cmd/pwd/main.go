package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
)

type CLIAction int

const (
	ActionRun CLIAction = iota
	ActionHelp
	ActionVersion
)

type ParsedOptions struct {
	Action CLIAction
	Mode   ResolutionMode
}

func printHelp() string {
	return "Usage:\n" +
		"pwd [OPTION]...\n" +
		"\n" +
		"Description:\n" +
		"Print the full filename of the current working directory.\n" +
		"\n" +
		"  -L, --logical\t\tuse PWD from environment, even if it contains symlinks (default)\n" +
		"  -P, --physical\tavoid all symlinks\n" +
		"  -h, --help\t\tdisplay this help and exit\n" +
		"  -v, --version\t\toutput version information and exit\n"
}

func printVersion() error {
	_, err := fmt.Println("pwd (coreutils-go) 0.1.0")
	return err
}

func parseArgs(args []string) (ParsedOptions, error) {
	var options ParsedOptions
	for _, arg := range args {
		if arg == "--" {
			break
		}

		switch arg {
		case "--help":
			options.Action = ActionHelp
			return options, nil
		case "--version":
			options.Action = ActionVersion
			return options, nil
		case "--logical":
			options.Mode = LogicalMode
			continue
		case "--physical":
			options.Mode = PhysicalMode
			continue
		}

		if strings.HasPrefix(arg, "--") {
			return options, fmt.Errorf("unrecognized option '%s'", arg)
		}

		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			for _, flagChar := range arg[1:] {
				switch flagChar {
				case 'h':
					options.Action = ActionHelp
					return options, nil
				case 'v':
					options.Action = ActionVersion
					return options, nil
				case 'L':
					options.Mode = LogicalMode
				case 'P':
					options.Mode = PhysicalMode
				default:
					return options, fmt.Errorf("invalid option -- '%c'", flagChar)
				}
			}
			continue
		}
	}
	return options, nil
}

func isBrokenPipe(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, syscall.EPIPE) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "broken pipe")
}

func run(args []string, stdout, stderr io.Writer) int {
	options, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "pwd: %v\n", err)
		return 1
	}

	switch options.Action {
	case ActionVersion:
		if _, err := fmt.Fprintln(stdout, "pwd (coreutils-go) 0.1.0"); err != nil && !isBrokenPipe(err) {
			return 1
		}
		return 0
	case ActionHelp:
		if _, err := fmt.Fprint(stdout, printHelp()); err != nil && !isBrokenPipe(err) {
			return 1
		}
		return 0
	case ActionRun:
		path, err := resolveWorkingDir(options.Mode)
		if err != nil {
			fmt.Fprintf(stderr, "pwd: %v\n", err)
			return 1
		}
		if _, err := fmt.Fprintln(stdout, path); err != nil && !isBrokenPipe(err) {
			return 1
		}
		return 0
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
