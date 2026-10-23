package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	showAllFlag                    bool
	numberNonBlankFlag             bool
	showNonPrintingAndShowEndsFlag bool
	showEndsFlag                   bool
	numberFlag                     bool
	squeezeBlankFlag               bool
	showNonPrintingAndShowTabsFlag bool
	showTabsFlag                   bool
	posixCompatFlag                bool
	showNonPrintingFlag            bool
)

const bufferSize = 64 * 1024 // 64 KB memory buffer for optimal I/O throughput

func runCatCmd(out io.Writer, r io.Reader) error {
	_, err := io.Copy(out, r)
	return err
}

func processSuperFlags() {
	// if -A or --show-all
	if showAllFlag {
		showNonPrintingFlag = true
		showEndsFlag = true
		showTabsFlag = true
	}
	if showNonPrintingAndShowEndsFlag {
		showNonPrintingFlag = true
		showEndsFlag = true
	}
	if showNonPrintingAndShowTabsFlag {
		showNonPrintingFlag = true
		showTabsFlag = true
	}
}

func useFlags() bool {
	res := numberNonBlankFlag
	res = res || showEndsFlag
	res = res || numberFlag
	res = res || squeezeBlankFlag
	res = res || showTabsFlag
	res = res || showNonPrintingFlag
	return res
}

func formatNonPrinting(line []byte) []byte {
	result := make([]byte, 0, len(line))
	for _, b := range line {
		switch {
		case b == '\n' || b == '\t':
			result = append(result, b)
		case b < 32:
			result = append(result, '^', b+64)
		case b < 127:
			result = append(result, b)
		case b == 127:
			result = append(result, '^', '?')
		case b < 160:
			result = append(result, 'M', '-', '^', b-128+64)
		case b < 255:
			result = append(result, 'M', '-', b-128)
		default: // b == 255
			result = append(result, 'M', '-', '^', '?')
		}
	}
	return result
}

func isBlank(line []byte) bool {
	return len(line) == 0 || (len(line) == 1 && line[0] == '\n') || (len(line) == 2 && line[0] == '\r' && line[1] == '\n')
}

var catCmd = &cobra.Command{
	Use:   "cat [OPTION]... [FILE]...",
	Short: "Concatenate FILE(s) to standard output.",
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion, err := cmd.Flags().GetBool("version"); err == nil && showVersion {
			if err := printVersion(); err != nil {
				os.Exit(1)
			}
			return
		}

		processSuperFlags()
		writer := bufio.NewWriter(cmd.OutOrStdout())
		defer writer.Flush()
		lineNumber := 1
		prevBlank := false
		hasError := false
		if len(args) == 0 {
			args = append(args, "-")
		}
		for _, filename := range args {
			var fileToClose *os.File
			var r io.Reader
			if filename == "-" {
				r = cmd.InOrStdin()
			} else {
				f, err := os.Open(filename)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "cat: %s: %v\n", filename, err)
					hasError = true
					continue
				}
				fileToClose = f
				r = f

				if stat, err := f.Stat(); err == nil && stat.IsDir() {
					fmt.Fprintf(cmd.ErrOrStderr(), "cat: %s: Is a directory\n", filename)
					fileToClose.Close()
					fileToClose = nil
					hasError = true
					continue
				}
			}

			if useFlags() {
				reader := bufio.NewReader(r)
				for {
					line, err := reader.ReadBytes('\n')
					if err == io.EOF && len(line) == 0 {
						break
					}
					if err != nil && err != io.EOF {
						fmt.Fprintf(cmd.ErrOrStderr(), "cat: %s: %v\n", filename, err)
						hasError = true
						break
					}

					if squeezeBlankFlag {
						if prevBlank && isBlank(line) {
							if err == io.EOF {
								break
							}
							continue
						}
						prevBlank = isBlank(line)
					}

					if numberNonBlankFlag {
						if isBlank(line) {
							writer.Write(line)
							if err == io.EOF {
								break
							}
							continue
						}
						fmt.Fprintf(writer, "%6d\t", lineNumber)
					} else if numberFlag {
						fmt.Fprintf(writer, "%6d\t", lineNumber)
					}
					if showNonPrintingFlag {
						line = formatNonPrinting(line)
					}
					if showTabsFlag {
						line = []byte(strings.ReplaceAll(string(line), "\t", "^I"))
					}
					if showEndsFlag {
						line = []byte(strings.ReplaceAll(string(line), "\n", "$\n"))
					}

					writer.Write(line)
					lineNumber++

					if err == io.EOF {
						break
					}
				}
			} else {
				if err := runCatCmd(writer, r); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "cat: %s: %v\n", filename, err)
					hasError = true
				}
			}
			if fileToClose != nil {
				fileToClose.Close()
			}
			fileToClose = nil
			writer.Flush()
		}

		if hasError {
			os.Exit(1)
		}
	},
}

func printHelp() string {
	return "Usage: cat [OPTION]... [FILE]...\n" +
		"Concatenate FILE(s) to standard output.\n" +
		"\n" +
		"With no FILE, or when FILE is -, read standard input.\n" +
		"\n" +
		"  -A, --show-all           equivalent to -vET\n" +
		"  -b, --number-nonblank    number nonempty output lines, overrides -n\n" +
		"  -e                       equivalent to -vE\n" +
		"  -E, --show-ends          display $ at end of each line\n" +
		"  -n, --number             number all output lines\n" +
		"  -s, --squeeze-blank      suppress repeated empty output lines\n" +
		"  -t                       equivalent to -vT\n" +
		"  -T, --show-tabs          display TAB characters as ^I\n" +
		"  -u                       (ignored)\n" +
		"  -v, --show-nonprinting   use ^ and M- notation, except for LFD and TAB\n" +
		"      --help        display this help and exit\n" +
		"      --version     output version information and exit\n" +
		"\n" +
		"Examples:\n" +
		"  cat f - g  Output f's contents, then standard input, then g's contents.\n" +
		"  cat        Copy standard input to standard output.\n"
}

func printVersion() error {
	_, err := fmt.Println("cat (coreutils-go) 0.1.0")
	return err
}

func init() {
	catCmd.SilenceUsage = true
	catCmd.SilenceErrors = true

	f := catCmd.Flags()
	f.BoolVarP(&showAllFlag, "show-all", "A", false, "equivalent to -vET")
	f.BoolVarP(&numberNonBlankFlag, "number-nonblank", "b", false, "number nonempty output lines, overrides -n")
	f.BoolVarS(&showNonPrintingAndShowEndsFlag, "vE", "e", false, "equivalent to -vE")
	f.BoolVarP(&showEndsFlag, "show-ends", "E", false, "display $ at end of each line")
	f.BoolVarP(&numberFlag, "number", "n", false, "number all output lines")
	f.BoolVarP(&squeezeBlankFlag, "squeeze-blank", "s", false, "suppress repeated empty output lines")
	f.BoolVarS(&showNonPrintingAndShowTabsFlag, "vT", "t", false, "equivalent to -vT")
	f.BoolVarP(&showTabsFlag, "show-tabs", "T", false, "display TAB characters as ^I")
	f.BoolVarS(&posixCompatFlag, "unbuffered", "u", false, "(ignored)")
	f.BoolVarP(&showNonPrintingFlag, "show-nonprinting", "v", false, "use ^ and M- notation, except for LFD and TAB")

	f.Bool("version", false, "output version information and exit")
	f.Bool("help", false, "display this help and exit")
	catCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprint(cmd.OutOrStdout(), printHelp())
	})
	catCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		if errors.Is(err, pflag.ErrHelp) {
			return fmt.Errorf("unknown shorthand flag: 'h' in -h")
		}
		return err
	})
	catCmd.SetUsageTemplate(printHelp())
}

func main() {
	if err := catCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
