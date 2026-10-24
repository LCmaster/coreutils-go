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
			"false [ignored command line arguments]",
			"false OPTION",
		},
		Description: "Exit with a status code indicating failure.",
		Options: []cli.HelpOption{
			cli.Opt('h', "help", "", "display this help and exit"),
			cli.Opt('v', "version", "", "output version information and exit"),
		},
	})
}

func getVersionStr() string {
	return version.Format("false")
}

func run(args []string, _ io.Reader, stdout, _ io.Writer) int {
	// GNU false specification:
	// --help and --version are only recognized when they are the sole argument.
	// When recognized, they display information and exit with status 0 (EXIT_SUCCESS).
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

	return 1
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
