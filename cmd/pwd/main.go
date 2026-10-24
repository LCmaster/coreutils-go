package main

import (
	"fmt"
	"io"
	"os"

	"github.com/LCmaster/coreutils-go/pkg/cli"
	"github.com/LCmaster/coreutils-go/pkg/ioutils"
)

func newCommand() *cli.Command {
	cmd := cli.NewCommand("pwd", "pwd [OPTION]...")
	cmd.SetDescription("Print the full filename of the current working directory.")
	cmd.EnableAutoShortVersion()

	// Default mode is Logical (physical = false)
	cmd.Options.Bools["physical"] = false

	cmd.FlagFunc('L', "logical", "use PWD from environment, even if it contains symlinks (default)", false, func(_ string) error {
		cmd.Options.Bools["physical"] = false
		return nil
	})
	cmd.FlagFunc('P', "physical", "avoid all symlinks", false, func(_ string) error {
		cmd.Options.Bools["physical"] = true
		return nil
	})

	cmd.Run = func(c *cli.Command, _ io.Reader, stdout, stderr io.Writer) error {
		path, err := resolveWorkingDir(c.Options.Bool("physical"))
		if err != nil {
			fmt.Fprintf(stderr, "pwd: %v\n", err)
			return err
		}
		if _, err := fmt.Fprintln(stdout, path); err != nil && !ioutils.IsBrokenPipe(err) {
			return err
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
