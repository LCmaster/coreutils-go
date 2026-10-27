package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/LCmaster/coreutils-go/pkg/cli"
	"github.com/LCmaster/coreutils-go/pkg/ioutils"
)

func newCommand() *cli.Command {
	cmd := cli.NewCommand("head")
	cmd.SetUsage("head [OPTION]... [FILE]...")
	cmd.DisableAutoShortHelp()
	cmd.DisableAutoShortVersion()
	cmd.SetDefaultOperands("-")
	cmd.SetDescription("Print the first 10 lines of each FILE to standard output.\n" +
		"With more than one FILE, precede each with a header giving the file name.\n\n" +
		"With no FILE, or when FILE is -, read standard input.")

	cmd.SetNotes("NUM may have a multiplier suffix:\n" +
		"b 512, kB 1000, K 1024, MB 1000*1000, M 1024*1024,\n" +
		"GB 1000*1000*1000, G 1024*1024*1024, and so on for T, P, E, Z, Y.\n" +
		"Binary prefixes can be used, too: KiB=K, MiB=M, and so on.")

	cmd.AddExample("head -n 5 file.txt     Output the first 5 lines of file.txt.")
	cmd.AddExample("head -c 1K file.txt    Output the first 1024 bytes of file.txt.")
	cmd.AddExample("head -q f1.txt f2.txt  Output contents of f1.txt and f2.txt without headers.")

	cmd.FlagArgFunc('c', "bytes", "[-]NUM", "print the first NUM bytes of each file;\nwith the leading '-', print all but the last\nNUM bytes of each file", func(number string) error {
		cmd.SetBool("byteMode", true)
		num, err := parseNum(number)
		if err != nil {
			return cli.ErrInvalidNumberOfBytes(number)
		}
		cmd.SetInt("bytes", num)
		return nil
	})
	cmd.FlagArgFunc('n', "lines", "[-]NUM", "print the first NUM lines instead of the first 10;\nwith the leading '-', print all but the last\nNUM lines of each file", func(number string) error {
		cmd.SetBool("byteMode", false)
		num, err := parseNum(number)
		if err != nil {
			return cli.ErrInvalidNumberOfLines(number)
		}
		cmd.SetInt("lines", num)
		return nil
	})
	cmd.FlagFunc('q', "quiet, silent", "never print headers giving file names", false, func(_ string) error {
		cmd.SetString("verbosity", "quiet")
		return nil
	})
	cmd.FlagFunc('v', "verbose", "always print headers giving file names", false, func(_ string) error {
		cmd.SetString("verbosity", "verbose")
		return nil
	})
	cmd.FlagFunc('z', "zero-terminated", "line delimiter is NUL, not newline", false, func(_ string) error {
		cmd.SetBool("zeroTerminated", true)
		return nil
	})

	cmd.EnableNumericShorthand("%s: invalid number of lines: '%s'", func(number string) error {
		cmd.SetBool("byteMode", false)
		num, err := parseNum(number)
		if err != nil || num <= 0 {
			return cli.ErrInvalidNumberOfLines(number)
		}
		cmd.SetInt("lines", num)
		return nil
	})

	cmd.Run = func(c *cli.Command, stdin io.Reader, stdout, stderr io.Writer) error {
		operands := c.Operands()
		printHeaders := false
		if c.String("verbosity") == "verbose" {
			printHeaders = true
		} else if c.String("verbosity") == "quiet" {
			printHeaders = false
		} else {
			printHeaders = len(operands) > 1
		}

		lineCount := 10
		if c.HasInt("lines") {
			lineCount = c.Int("lines")
		}
		byteMode := c.Bool("byteMode")
		byteCount := c.Int("bytes")

		delimiter := byte('\n')
		if c.Bool("zeroTerminated") {
			delimiter = 0
		}

		writer := bufio.NewWriterSize(stdout, ioutils.DefaultBufferSize)
		defer writer.Flush()

		hasError := false
		filesProcessed := 0

		for _, filename := range operands {
			r, fileToClose, err := ioutils.OpenInput(filename, stdin)
			if err != nil {
				if printHeaders {
					if filesProcessed > 0 {
						fmt.Fprintln(writer)
					}
					fmt.Fprintf(writer, "==> %s <==\n", filename)
					if fErr := writer.Flush(); fErr != nil && ioutils.IsBrokenPipe(fErr) {
						break
					}
					filesProcessed++
				}
				c.Errorf(stderr, "%v", cli.ErrCannotOpenForReading(filename, err))
				hasError = true
				continue
			}

			if printHeaders {
				if filesProcessed > 0 {
					fmt.Fprintln(writer)
				}
				if filename == "-" {
					fmt.Fprintln(writer, "==> standard input <==")
				} else {
					fmt.Fprintf(writer, "==> %s <==\n", filename)
				}
				if err := writer.Flush(); err != nil && ioutils.IsBrokenPipe(err) {
					_ = fileToClose.Close()
					break
				}
			}

			err = headFile(r, writer, byteMode, byteCount, lineCount, delimiter)
			_ = fileToClose.Close()

			if err != nil {
				if ioutils.IsBrokenPipe(err) {
					break
				}
				c.Errorf(stderr, "%v", cli.ErrReadingFile(filename, err))
				hasError = true
			}

			if err := writer.Flush(); err != nil && ioutils.IsBrokenPipe(err) {
				break
			}
			filesProcessed++
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
