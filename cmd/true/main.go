package main

import (
	"io"
	"os"

	"github.com/LCmaster/coreutils-go/pkg/cli"
)

func newCommand() *cli.Command {
	cmd := cli.NewCommand("true")
	cmd.Helper.SetUsage("true [ignored command line arguments]", "true OPTION")
	cmd.Helper.SetDescription("Exit with a status code indicating success.")
	cmd.EnableAutoShortVersion()
	cmd.EnableAutoShortHelp()
	cmd.Run = func(_ *cli.Command, _ io.Reader, _, _ io.Writer) error {
		return nil
	}

	return cmd
}

func main() {
	if err := newCommand().Execute(); err != nil {
		os.Exit(1)
	}
}



