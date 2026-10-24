package main

import (
	"errors"
	"io"
	"os"

	"github.com/LCmaster/coreutils-go/pkg/cli"
)

func newCommand() *cli.Command {
	cmd := cli.NewCommand("false")
	cmd.SetUsage("false [ignored command line arguments]", "false OPTION")
	cmd.SetDescription("Exit with a status code indicating failure.")
	cmd.EnableAutoShortVersion()
	cmd.EnableAutoShortHelp()
	cmd.Run = func(_ *cli.Command, _ io.Reader, _, _ io.Writer) error {
		return errors.New("exit status 1")
	}
	return cmd
}

func main() {
	if err := newCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
