package main

import (
	"fmt"
	"io"
	"os"

	"github.com/LCmaster/coreutils-go/pkg/cli"
	"github.com/LCmaster/coreutils-go/pkg/ioutils"
	"github.com/LCmaster/coreutils-go/pkg/version"
)

func getHelpStr() string {
	return cli.FormatHelp(cli.HelpConfig{
		Usage: []string{
			"true [ignored command line arguments]",
			"true OPTION",
		},
		Description: "Exit with a status code indicating success.",
		Options: []cli.HelpOption{
			cli.Opt('h', "help", "", "display this help and exit"),
			cli.Opt('v', "version", "", "output version information and exit"),
		},
	})
}

func getVersionStr() string {
	return version.Format("true")
}

func run(args []string, _ io.Reader, stdout, _ io.Writer) int {
	// GNU true specification:
	// --help and --version are only recognized when they are the sole argument.
	if len(args) == 1 {
		switch args[0] {
		case "-h", "--help":
			if _, err := fmt.Fprint(stdout, getHelpStr()); err != nil && !ioutils.IsBrokenPipe(err) {
				return 1
			}
			return 0
		case "-v", "--version":
			if _, err := fmt.Fprintln(stdout, getVersionStr()); err != nil && !ioutils.IsBrokenPipe(err) {
				return 1
			}
			return 0
		}
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
