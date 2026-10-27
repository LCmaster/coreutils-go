package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/LCmaster/coreutils-go/pkg/cli"
	"github.com/LCmaster/coreutils-go/pkg/ioutils"
)

type wcCounts struct {
	lines      int64
	words      int64
	chars      int64
	bytes      int64
	maxLineLen int64
}

func (c *wcCounts) add(other wcCounts) {
	c.lines += other.lines
	c.words += other.words
	c.chars += other.chars
	c.bytes += other.bytes
	if other.maxLineLen > c.maxLineLen {
		c.maxLineLen = other.maxLineLen
	}
}

func countReader(r io.Reader) (wcCounts, error) {
	reader := bufio.NewReaderSize(r, ioutils.DefaultBufferSize)
	var counts wcCounts
	var curLineLen int64
	inWord := false

	for {
		ru, size, err := reader.ReadRune()
		if size > 0 {
			counts.bytes += int64(size)
			counts.chars++

			switch ru {
			case '\n':
				counts.lines++
				if curLineLen > counts.maxLineLen {
					counts.maxLineLen = curLineLen
				}
				curLineLen = 0
			case '\r':
				if curLineLen > counts.maxLineLen {
					counts.maxLineLen = curLineLen
				}
				curLineLen = 0
			case '\t':
				curLineLen += 8 - (curLineLen % 8)
			default:
				if unicode.IsPrint(ru) {
					curLineLen++
				}
			}

			if unicode.IsSpace(ru) {
				inWord = false
			} else if !inWord {
				inWord = true
				counts.words++
			}
		}

		if err == io.EOF {
			if curLineLen > counts.maxLineLen {
				counts.maxLineLen = curLineLen
			}
			break
		}
		if err != nil {
			return counts, err
		}
	}

	return counts, nil
}

func readFiles0From(source string, stdin io.Reader) ([]string, error) {
	var r io.Reader = stdin
	if source != "-" {
		f, err := os.Open(source)
		if err != nil {
			return nil, fmt.Errorf("cannot open '%s' for reading: %v", source, err)
		}
		defer f.Close()
		r = f
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	parts := bytes.Split(data, []byte{0})
	if len(parts) > 0 && len(parts[len(parts)-1]) == 0 {
		parts = parts[:len(parts)-1]
	}

	files := make([]string, 0, len(parts))
	for _, p := range parts {
		files = append(files, string(p))
	}
	return files, nil
}

func formatRow(c wcCounts, label string, width int, showLines, showWords, showChars, showBytes, showMaxLine bool) string {
	var vals []int64
	if showLines {
		vals = append(vals, c.lines)
	}
	if showWords {
		vals = append(vals, c.words)
	}
	if showChars {
		vals = append(vals, c.chars)
	}
	if showBytes {
		vals = append(vals, c.bytes)
	}
	if showMaxLine {
		vals = append(vals, c.maxLineLen)
	}

	var sb strings.Builder
	for i, v := range vals {
		if i > 0 {
			sb.WriteByte(' ')
		}
		fmt.Fprintf(&sb, "%*d", width, v)
	}
	if label != "" {
		sb.WriteByte(' ')
		sb.WriteString(label)
	}
	sb.WriteByte('\n')
	return sb.String()
}

func computeColWidth(total wcCounts, numFiles, numCols int, showLines, showWords, showChars, showBytes, showMaxLine bool) int {
	if numFiles == 1 && numCols == 1 {
		return 0
	}
	var maxVal int64
	if showLines && total.lines > maxVal {
		maxVal = total.lines
	}
	if showWords && total.words > maxVal {
		maxVal = total.words
	}
	if showChars && total.chars > maxVal {
		maxVal = total.chars
	}
	if showBytes && total.bytes > maxVal {
		maxVal = total.bytes
	}
	if showMaxLine && total.maxLineLen > maxVal {
		maxVal = total.maxLineLen
	}
	digits := len(strconv.FormatInt(maxVal, 10))
	if numCols > 1 && numFiles == 1 && digits < 7 {
		return 7
	}
	return digits
}

func newCommand() *cli.Command {
	cmd := cli.NewCommand("wc")
	cmd.SetUsage("wc [OPTION]... [FILE]...", "wc [OPTION]... --files0-from=F")
	cmd.DisableAutoShortHelp()
	cmd.DisableAutoShortVersion()
	cmd.SetDefaultOperands("-")
	cmd.SetDescription("Print newline, word, and byte counts for each FILE, and a total line if\n" +
		"more than one FILE is specified. A word is a non-empty-length sequence of\n" +
		"characters delimited by white space.\n\n" +
		"With no FILE, or when FILE is -, read standard input.")

	cmd.FlagFunc('c', "bytes", "print the byte counts", false, func(_ string) error {
		cmd.SetBool("bytes", true)
		return nil
	})
	cmd.FlagFunc('m', "chars", "print the character counts", false, func(_ string) error {
		cmd.SetBool("chars", true)
		return nil
	})
	cmd.FlagFunc('l', "lines", "print the newline counts", false, func(_ string) error {
		cmd.SetBool("lines", true)
		return nil
	})
	cmd.FlagArgFunc(0, "files0-from", "F", "read input from the files specified by\nNUL-terminated names in file F;\nif F is - then read names from standard input", func(file string) error {
		cmd.SetString("files0-from", file)
		return nil
	})
	cmd.FlagFunc('L', "max-line-length", "print the maximum display width", false, func(_ string) error {
		cmd.SetBool("max-line-length", true)
		return nil
	})
	cmd.FlagFunc('w', "words", "print the word counts", false, func(_ string) error {
		cmd.SetBool("words", true)
		return nil
	})

	cmd.OnError = func(c *cli.Command, stderr io.Writer, err error) int {
		if stderr != nil {
			fmt.Fprintf(stderr, "%v\nTry '%s --help' for more information.\n", err, c.Name)
		}
		return 1
	}

	cmd.Run = func(c *cli.Command, stdin io.Reader, stdout, stderr io.Writer) error {
		showLines := c.Bool("lines")
		showWords := c.Bool("words")
		showChars := c.Bool("chars")
		showBytes := c.Bool("bytes")
		showMaxLine := c.Bool("max-line-length")

		if !showLines && !showWords && !showChars && !showBytes && !showMaxLine {
			showLines, showWords, showBytes = true, true, true
		}

		numCols := 0
		for _, b := range []bool{showLines, showWords, showChars, showBytes, showMaxLine} {
			if b {
				numCols++
			}
		}

		operands := c.Operands()
		if c.HasString("files0-from") {
			files0From := c.String("files0-from")
			if files0From == "" {
				fmt.Fprintf(stderr, "wc: cannot open '' for reading: No such file or directory\n")
				return cli.ErrExitFailure
			}
			var err error
			operands, err = readFiles0From(files0From, stdin)
			if err != nil {
				fmt.Fprintf(stderr, "wc: %v\n", err)
				return cli.ErrExitFailure
			}
		}

		writer := bufio.NewWriterSize(stdout, ioutils.DefaultBufferSize)
		defer writer.Flush()

		var total wcCounts
		hasError := false

		type fileResult struct {
			counts wcCounts
			label  string
		}
		var results []fileResult

		for _, filename := range operands {
			if filename == "" {
				fmt.Fprintf(stderr, "wc: invalid zero-length file name\n")
				hasError = true
				continue
			}

			var r io.Reader = stdin
			var fileToClose io.Closer

			if filename != "-" {
				f, err := os.Open(filename)
				if err != nil {
					fmt.Fprintf(stderr, "wc: %s: %v\n", filename, err)
					hasError = true
					continue
				}
				r = f
				fileToClose = f
			}

			counts, err := countReader(r)
			if fileToClose != nil {
				fileToClose.Close()
			}

			if err != nil {
				fmt.Fprintf(stderr, "wc: %s: %v\n", filename, err)
				hasError = true
				continue
			}

			total.add(counts)
			label := filename
			if filename == "-" {
				label = ""
			}
			results = append(results, fileResult{counts: counts, label: label})
		}

		width := computeColWidth(total, len(operands), numCols, showLines, showWords, showChars, showBytes, showMaxLine)
		for _, res := range results {
			row := formatRow(res.counts, res.label, width, showLines, showWords, showChars, showBytes, showMaxLine)
			if _, err := writer.WriteString(row); err != nil {
				if ioutils.IsBrokenPipe(err) {
					return nil
				}
				return err
			}
		}

		if len(operands) > 1 {
			row := formatRow(total, "total", width, showLines, showWords, showChars, showBytes, showMaxLine)
			if _, err := writer.WriteString(row); err != nil && !ioutils.IsBrokenPipe(err) {
				return err
			}
		}

		if err := writer.Flush(); err != nil && !ioutils.IsBrokenPipe(err) {
			return err
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
