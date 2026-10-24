package main

import (
	"fmt"
	"io"
	"os"

	"github.com/LCmaster/coreutils-go/pkg/cli"
	"github.com/LCmaster/coreutils-go/pkg/ioutils"
	"github.com/LCmaster/coreutils-go/pkg/version"
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

func getHelpStr() string {
	return cli.FormatHelp(cli.HelpConfig{
		Usage:       []string{"pwd [OPTION]..."},
		Description: "Print the full filename of the current working directory.",
		Options: []cli.HelpOption{
			cli.Opt('L', "logical", "", "use PWD from environment, even if it contains symlinks (default)"),
			cli.Opt('P', "physical", "", "avoid all symlinks"),
			cli.Opt('h', "help", "", "display this help and exit"),
			cli.Opt('v', "version", "", "output version information and exit"),
		},
	})
}

func getVersionStr() string {
	return version.Format("pwd")
}

func parseArgs(args []string) (ParsedOptions, error) {
	opts := ParsedOptions{
		Action: ActionRun,
		Mode:   LogicalMode,
	}

	p := cli.NewParser("pwd", "pwd [OPTION]...")
	p.FlagFunc('L', "logical", "use PWD from environment, even if it contains symlinks (default)", false, func(_ string) error {
		opts.Mode = LogicalMode
		return nil
	})
	p.FlagFunc('P', "physical", "avoid all symlinks", false, func(_ string) error {
		opts.Mode = PhysicalMode
		return nil
	})
	p.FlagFunc('v', "", "output version information and exit", false, func(_ string) error {
		opts.Action = ActionVersion
		return nil
	})

	_, action, err := p.Parse(args)
	if err != nil {
		return opts, err
	}
	switch action {
	case cli.ActionHelp:
		opts.Action = ActionHelp
	case cli.ActionVersion:
		opts.Action = ActionVersion
	}
	return opts, nil
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	options, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	switch options.Action {
	case ActionHelp:
		if _, err := fmt.Fprint(stdout, getHelpStr()); err != nil && !ioutils.IsBrokenPipe(err) {
			return 1
		}
		return 0
	case ActionVersion:
		if _, err := fmt.Fprintln(stdout, getVersionStr()); err != nil && !ioutils.IsBrokenPipe(err) {
			return 1
		}
		return 0
	case ActionRun:
		path, err := resolveWorkingDir(options.Mode)
		if err != nil {
			fmt.Fprintf(stderr, "pwd: %v\n", err)
			return 1
		}
		if _, err := fmt.Fprintln(stdout, path); err != nil && !ioutils.IsBrokenPipe(err) {
			return 1
		}
		return 0
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
