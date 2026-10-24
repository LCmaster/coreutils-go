package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LCmaster/coreutils-go/pkg/cli"
	"github.com/LCmaster/coreutils-go/pkg/ioutils"
	"github.com/LCmaster/coreutils-go/pkg/version"
)

const bufferSize = ioutils.DefaultBufferSize // Memory buffer for optimal I/O throughput

func getHelpStr() string {
	return cli.FormatHelp(cli.HelpConfig{
		Usage: []string{
			"yes [STRING]...",
			"yes OPTION",
		},
		Description: "Repeatedly output a line with all specified STRING(s), or 'y'.",
		Options: []cli.HelpOption{
			cli.Opt('h', "help", "", "display this help and exit"),
			cli.Opt('v', "version", "", "output version information and exit"),
		},
	})
}

func getVersionStr() string {
	return version.Format("yes")
}

// buildBuffer creates a reusable chunk buffer pre-filled with line repetitions
// up to bufferSize to minimize system call overhead.
func buildBuffer(line []byte) []byte {
	if len(line) >= bufferSize {
		return line
	}

	repeats := bufferSize / len(line)
	if repeats == 0 {
		repeats = 1
	}

	return bytes.Repeat(line, repeats)
}

// runYes continuously writes the given arguments or 'y' to w until an error occurs.
func runYes(w io.Writer, args []string) error {
	var line []byte
	if len(args) == 0 {
		line = []byte("y\n")
	} else {
		line = []byte(strings.Join(args, " ") + "\n")
	}

	buf := buildBuffer(line)
	for {
		if _, err := w.Write(buf); err != nil {
			return err
		}
	}
}

func run(args []string, _ io.Reader, stdout, stderr io.Writer) int {
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

	err := runYes(stdout, args)
	if err != nil && !ioutils.IsBrokenPipe(err) {
		fmt.Fprintf(stderr, "yes: write error: %v\n", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
