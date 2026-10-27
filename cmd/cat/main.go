package main

import (
	"bufio"
	"io"
	"os"

	"github.com/LCmaster/coreutils-go/pkg/cli"
	"github.com/LCmaster/coreutils-go/pkg/ioutils"
)

func newCommand() *cli.Command {
	cmd := cli.NewCommand("cat", "cat [OPTION]... [FILE]...")
	cmd.DisableAutoShortHelp()
	cmd.SetDefaultOperands("-")
	cmd.SetDescription("Concatenate FILE(s) to standard output.\n" +
		"With no FILE, or when FILE is -, read standard input.")
	cmd.AddExample("cat f - g  Output f's contents, then standard input, then g's contents.")
	cmd.AddExample("cat        Copy standard input to standard output.")

	cmd.FlagFunc('A', "show-all", "equivalent to -vET", false, func(_ string) error {
		cmd.SetBool("ShowNonPrinting", true)
		cmd.SetBool("ShowEnds", true)
		cmd.SetBool("ShowTabs", true)
		return nil
	})
	cmd.FlagFunc('b', "number-nonblank", "number nonempty output lines, overrides -n", false, func(_ string) error {
		cmd.SetBool("NumberNonBlank", true)
		return nil
	})
	cmd.FlagFunc('e', "", "equivalent to -vE", false, func(_ string) error {
		cmd.SetBool("ShowNonPrinting", true)
		cmd.SetBool("ShowEnds", true)
		return nil
	})
	cmd.FlagFunc('E', "show-ends", "display $ at end of each line", false, func(_ string) error {
		cmd.SetBool("ShowEnds", true)
		return nil
	})
	cmd.FlagFunc('n', "number", "number all output lines", false, func(_ string) error {
		cmd.SetBool("Number", true)
		return nil
	})
	cmd.FlagFunc('s', "squeeze-blank", "suppress repeated empty output lines", false, func(_ string) error {
		cmd.SetBool("SqueezeBlank", true)
		return nil
	})
	cmd.FlagFunc('t', "", "equivalent to -vT", false, func(_ string) error {
		cmd.SetBool("ShowNonPrinting", true)
		cmd.SetBool("ShowTabs", true)
		return nil
	})
	cmd.FlagFunc('T', "show-tabs", "display TAB characters as ^I", false, func(_ string) error {
		cmd.SetBool("ShowTabs", true)
		return nil
	})
	cmd.FlagFunc('u', "", "(ignored)", false, func(_ string) error {
		cmd.SetBool("Unbuffered", true)
		return nil
	})
	cmd.FlagFunc('v', "show-nonprinting", "use ^ and M- notation, except for LFD and TAB", false, func(_ string) error {
		cmd.SetBool("ShowNonPrinting", true)
		return nil
	})

	cmd.Run = func(c *cli.Command, stdin io.Reader, stdout, stderr io.Writer) error {
		opts := CatOptions{
			NumberNonBlank:  c.Bool("NumberNonBlank"),
			Number:          c.Bool("Number"),
			ShowEnds:        c.Bool("ShowEnds"),
			ShowTabs:        c.Bool("ShowTabs"),
			ShowNonPrinting: c.Bool("ShowNonPrinting"),
			SqueezeBlank:    c.Bool("SqueezeBlank"),
			Unbuffered:      c.Bool("Unbuffered"),
		}
		if opts.NumberNonBlank {
			opts.Number = false
		}
		pipeline := newFormatPipeline(opts)
		writer := bufio.NewWriterSize(stdout, bufferSize)
		defer writer.Flush()

		state := StreamState{LineNumber: 1}
		hasError := false

		for _, filename := range c.Operands() {
			err := processFile(filename, stdin, stdout, opts, &state, pipeline, writer)
			if err != nil {
				if ioutils.IsBrokenPipe(err) {
					break
				}
				c.FileError(stderr, filename, err)
				hasError = true
			}
			if err := writer.Flush(); err != nil && ioutils.IsBrokenPipe(err) {
				break
			}
		}

		if hasError {
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
