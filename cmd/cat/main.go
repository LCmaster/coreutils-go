package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func parseOptions(cmd *cobra.Command) Options {
	f := cmd.Flags()

	showAll, _ := f.GetBool("show-all")
	showNonPrinting, _ := f.GetBool("show-nonprinting")
	showEnds, _ := f.GetBool("show-ends")
	showTabs, _ := f.GetBool("show-tabs")
	numberNonBlank, _ := f.GetBool("number-nonblank")
	number, _ := f.GetBool("number")
	squeezeBlank, _ := f.GetBool("squeeze-blank")
	unbuffered, _ := f.GetBool("unbuffered")
	eFlag, _ := f.GetBool("e")
	tFlag, _ := f.GetBool("t")

	return Options{
		NumberNonBlank:  numberNonBlank,
		Number:          number && !numberNonBlank,
		SqueezeBlank:    squeezeBlank,
		ShowEnds:        showEnds || showAll || eFlag,
		ShowTabs:        showTabs || showAll || tFlag,
		ShowNonPrinting: showNonPrinting || showAll || eFlag || tFlag,
		Unbuffered:      unbuffered,
	}
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

		opts := parseOptions(cmd)
		pipeline := newFormatPipeline(opts)
		writer := bufio.NewWriterSize(cmd.OutOrStdout(), bufferSize)
		defer writer.Flush()

		if len(args) == 0 {
			args = []string{"-"}
		}

		state := StreamState{LineNumber: 1}
		hasError := false

		for _, filename := range args {
			err := processFile(filename, cmd.InOrStdin(), cmd.OutOrStdout(), opts, &state, pipeline, writer)
			if err != nil {
				if isBrokenPipe(err) {
					break
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "cat: %s: %v\n", filename, err)
				hasError = true
			}
			if err := writer.Flush(); err != nil && isBrokenPipe(err) {
				break
			}
		}

		if hasError {
			os.Exit(1)
		}
	},
}

func printHelp() string {
	helpMessage := "" +
		"Usage:\n" +
		"cat [OPTION]... [FILE]...\n" +
		"\n" +
		"Description:\n" +
		"Concatenate FILE(s) to standard output.\n" +
		"With no FILE, or when FILE is -, read standard input.\n" +
		"\n" +
		"  -A, --show-all\t\tequivalent to -vET\n" +
		"  -b, --number-nonblank\t\tnumber nonempty output lines, overrides -n\n" +
		"  -e\t\t\t\tequivalent to -vE\n" +
		"  -E, --show-ends\t\tdisplay $ at end of each line\n" +
		"  -n, --number\t\t\tnumber all output lines\n" +
		"  -s, --squeeze-blank\t\tsuppress repeated empty output lines\n" +
		"  -t\t\t\t\tequivalent to -vT\n" +
		"  -T, --show-tabs\t\tdisplay TAB characters as ^I\n" +
		"  -u\t\t\t\t(ignored)\n" +
		"  -v, --show-nonprinting\tuse ^ and M- notation, except for LFD and TAB\n" +
		"      --help\t\t\tdisplay this help and exit\n" +
		"      --version\t\t\toutput version information and exit\n" +
		"\n" +
		"Examples:\n" +
		"  cat f - g  Output f's contents, then standard input, then g's contents.\n" +
		"  cat        Copy standard input to standard output.\n"

	return helpMessage
}

func printVersion() error {
	_, err := fmt.Println("cat (coreutils-go) 0.1.0")
	return err
}

func init() {
	catCmd.SilenceUsage = true
	catCmd.SilenceErrors = true

	f := catCmd.Flags()
	f.BoolP("show-all", "A", false, "equivalent to -vET")
	f.BoolP("number-nonblank", "b", false, "number nonempty output lines, overrides -n")
	f.BoolS("e", "e", false, "equivalent to -vE")
	f.BoolP("show-ends", "E", false, "display $ at end of each line")
	f.BoolP("number", "n", false, "number all output lines")
	f.BoolP("squeeze-blank", "s", false, "suppress repeated empty output lines")
	f.BoolS("t", "t", false, "equivalent to -vT")
	f.BoolP("show-tabs", "T", false, "display TAB characters as ^I")
	f.BoolS("unbuffered", "u", false, "(ignored)")
	f.BoolP("show-nonprinting", "v", false, "use ^ and M- notation, except for LFD and TAB")

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
