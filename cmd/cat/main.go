package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type CLIAction int

const (
	ActionRun CLIAction = iota
	ActionHelp
	ActionVersion
)

func parseArgs(args []string) (Options, []string, CLIAction, error) {
	var opts Options
	var files []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--" {
			files = append(files, args[i+1:]...)
			break
		}

		if arg == "-" || !strings.HasPrefix(arg, "-") {
			files = append(files, arg)
			continue
		}

		switch arg {
		case "--help":
			return opts, nil, ActionHelp, nil
		case "--version":
			return opts, nil, ActionVersion, nil
		case "--show-all":
			opts.ShowNonPrinting = true
			opts.ShowEnds = true
			opts.ShowTabs = true
		case "--number-nonblank":
			opts.NumberNonBlank = true
		case "--show-ends":
			opts.ShowEnds = true
		case "--number":
			opts.Number = true
		case "--squeeze-blank":
			opts.SqueezeBlank = true
		case "--show-tabs":
			opts.ShowTabs = true
		case "--show-nonprinting":
			opts.ShowNonPrinting = true
		default:
			if strings.HasPrefix(arg, "--") {
				return opts, nil, ActionRun, fmt.Errorf("unrecognized option '%s'", arg)
			}

			for _, flagChar := range arg[1:] {
				switch flagChar {
				case 'A':
					opts.ShowNonPrinting = true
					opts.ShowEnds = true
					opts.ShowTabs = true
				case 'b':
					opts.NumberNonBlank = true
				case 'e':
					opts.ShowNonPrinting = true
					opts.ShowEnds = true
				case 'E':
					opts.ShowEnds = true
				case 'n':
					opts.Number = true
				case 's':
					opts.SqueezeBlank = true
				case 't':
					opts.ShowNonPrinting = true
					opts.ShowTabs = true
				case 'T':
					opts.ShowTabs = true
				case 'u':
					opts.Unbuffered = true
				case 'v':
					opts.ShowNonPrinting = true
				default:
					return opts, nil, ActionRun, fmt.Errorf("invalid option -- '%c'", flagChar)
				}
			}
		}
	}

	if opts.NumberNonBlank {
		opts.Number = false
	}

	if len(files) == 0 {
		files = []string{"-"}
	}

	return opts, files, ActionRun, nil
}

func printHelp() string {
	helpMessage := "" +
		"Usage:\n" +
		"cat [OPTION]... [FILE]...\n" +
		"\n" +
		"Description:\n" +
		"Concatenate FILE(s) to standard output.\n" +
		"With no FILE, or when FILE is -, read standard input.\n" +
		"\n" +
		"  -A, --show-all\t\tequivalent to -vET\n" +
		"  -b, --number-nonblank\t\tnumber nonempty output lines, overrides -n\n" +
		"  -e\t\t\t\tequivalent to -vE\n" +
		"  -E, --show-ends\t\tdisplay $ at end of each line\n" +
		"  -n, --number\t\t\tnumber all output lines\n" +
		"  -s, --squeeze-blank\t\tsuppress repeated empty output lines\n" +
		"  -t\t\t\t\tequivalent to -vT\n" +
		"  -T, --show-tabs\t\tdisplay TAB characters as ^I\n" +
		"  -u\t\t\t\t(ignored)\n" +
		"  -v, --show-nonprinting\tuse ^ and M- notation, except for LFD and TAB\n" +
		"      --help\t\t\tdisplay this help and exit\n" +
		"      --version\t\t\toutput version information and exit\n" +
		"\n" +
		"Examples:\n" +
		"  cat f - g  Output f's contents, then standard input, then g's contents.\n" +
		"  cat        Copy standard input to standard output.\n"

	return helpMessage
}

func printVersion(w io.Writer) error {
	_, err := fmt.Fprintln(w, "cat (coreutils-go) 0.1.0")
	return err
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	opts, files, action, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "cat: %v\nTry 'cat --help' for more information.\n", err)
		return 1
	}

	switch action {
	case ActionVersion:
		_ = printVersion(stdout)
		return 0
	case ActionHelp:
		fmt.Fprint(stdout, printHelp())
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
				if isBrokenPipe(err) {
					break
				}
				fmt.Fprintf(stderr, "cat: %s: %v\n", filename, err)
				hasError = true
			}
			if err := writer.Flush(); err != nil && isBrokenPipe(err) {
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
