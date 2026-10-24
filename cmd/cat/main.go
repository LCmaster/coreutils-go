package main

import (
	"bufio"
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

func parseArgs(args []string) (Options, []string, CLIAction, error) {
	var opts Options
	p := cli.NewParser("cat", "cat [OPTION]... [FILE]...")
	p.DisableAutoShortHelp()
	p.SetDefaultOperands("-")

	p.FlagFunc('A', "show-all", "equivalent to -vET", false, func(_ string) error {
		opts.ShowNonPrinting = true
		opts.ShowEnds = true
		opts.ShowTabs = true
		return nil
	})
	p.FlagFunc('b', "number-nonblank", "number nonempty output lines, overrides -n", false, func(_ string) error {
		opts.NumberNonBlank = true
		return nil
	})
	p.FlagFunc('e', "", "equivalent to -vE", false, func(_ string) error {
		opts.ShowNonPrinting = true
		opts.ShowEnds = true
		return nil
	})
	p.FlagFunc('E', "show-ends", "display $ at end of each line", false, func(_ string) error {
		opts.ShowEnds = true
		return nil
	})
	p.FlagFunc('n', "number", "number all output lines", false, func(_ string) error {
		opts.Number = true
		return nil
	})
	p.FlagFunc('s', "squeeze-blank", "suppress repeated empty output lines", false, func(_ string) error {
		opts.SqueezeBlank = true
		return nil
	})
	p.FlagFunc('t', "", "equivalent to -vT", false, func(_ string) error {
		opts.ShowNonPrinting = true
		opts.ShowTabs = true
		return nil
	})
	p.FlagFunc('T', "show-tabs", "display TAB characters as ^I", false, func(_ string) error {
		opts.ShowTabs = true
		return nil
	})
	p.FlagFunc('u', "", "(ignored)", false, func(_ string) error {
		opts.Unbuffered = true
		return nil
	})
	p.FlagFunc('v', "show-nonprinting", "use ^ and M- notation, except for LFD and TAB", false, func(_ string) error {
		opts.ShowNonPrinting = true
		return nil
	})

	files, action, err := p.Parse(args)
	if err != nil {
		return opts, nil, ActionRun, err
	}

	var cliAction CLIAction = ActionRun
	switch action {
	case cli.ActionHelp:
		cliAction = ActionHelp
	case cli.ActionVersion:
		cliAction = ActionVersion
	}

	if opts.NumberNonBlank {
		opts.Number = false
	}

	return opts, files, cliAction, nil
}

func getHelpStr() string {
	return cli.FormatHelp(cli.HelpConfig{
		Usage: []string{"cat [OPTION]... [FILE]..."},
		Description: "Concatenate FILE(s) to standard output.\n" +
			"With no FILE, or when FILE is -, read standard input.",
		Options: []cli.HelpOption{
			cli.Opt('A', "show-all", "", "equivalent to -vET"),
			cli.Opt('b', "number-nonblank", "", "number nonempty output lines, overrides -n"),
			cli.Opt('e', "", "", "equivalent to -vE"),
			cli.Opt('E', "show-ends", "", "display $ at end of each line"),
			cli.Opt('n', "number", "", "number all output lines"),
			cli.Opt('s', "squeeze-blank", "", "suppress repeated empty output lines"),
			cli.Opt('t', "", "", "equivalent to -vT"),
			cli.Opt('T', "show-tabs", "", "display TAB characters as ^I"),
			cli.Opt('u', "", "", "(ignored)"),
			cli.Opt('v', "show-nonprinting", "", "use ^ and M- notation, except for LFD and TAB"),
			cli.Opt(0, "help", "", "display this help and exit"),
			cli.Opt(0, "version", "", "output version information and exit"),
		},
		Examples: []string{
			"cat f - g  Output f's contents, then standard input, then g's contents.",
			"cat        Copy standard input to standard output.",
		},
	})
}

func getVersionStr() string {
	return version.Format("cat")
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	opts, files, action, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "%v\nTry 'cat --help' for more information.\n", err)
		return 1
	}

	switch action {
	case ActionVersion:
		if _, err := fmt.Fprintln(stdout, getVersionStr()); err != nil && !ioutils.IsBrokenPipe(err) {
			return 1
		}
		return 0
	case ActionHelp:
		if _, err := fmt.Fprint(stdout, getHelpStr()); err != nil && !ioutils.IsBrokenPipe(err) {
			return 1
		}
		return 0
	case ActionRun:
		pipeline := newFormatPipeline(opts)
		writer := bufio.NewWriterSize(stdout, bufferSize)
		defer writer.Flush()

		state := StreamState{LineNumber: 1}
		hasError := false

		for _, filename := range files {
			err := processFile(filename, stdin, stdout, opts, &state, pipeline, writer)
			if err != nil {
				if ioutils.IsBrokenPipe(err) {
					break
				}
				fmt.Fprintf(stderr, "cat: %s: %v\n", filename, err)
				hasError = true
			}
			if err := writer.Flush(); err != nil && ioutils.IsBrokenPipe(err) {
				break
			}
		}

		if hasError {
			return 1
		}
		return 0
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
