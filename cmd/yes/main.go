package main

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/LCmaster/coreutils-go/pkg/cli"
	"github.com/LCmaster/coreutils-go/pkg/ioutils"
)

const bufferSize = ioutils.DefaultBufferSize // Memory buffer for optimal I/O throughput

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

func newCommand() *cli.Command {
	cmd := cli.NewCommand("yes")
	cmd.SetUsage("yes [STRING]...", "yes OPTION")
	cmd.SetDescription("Repeatedly output a line with all specified STRING(s), or 'y'.")
	cmd.EnableSoleArgumentHelpVersion()
	cmd.EnableAutoShortHelp()
	cmd.EnableAutoShortVersion()

	cmd.Run = func(c *cli.Command, _ io.Reader, stdout, stderr io.Writer) error {
		err := runYes(stdout, c.Operands())
		if err != nil && !ioutils.IsBrokenPipe(err) {
			c.Errorf(stderr, "%v", cli.ErrWrite(err))
			return cli.ErrExitFailure
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
