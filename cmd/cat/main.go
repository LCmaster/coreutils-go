package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

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

func parseOptions() (Options, bool) {
	opts := Options{
		NumberNonBlank:  numberNonBlankFlag,
		ShowEnds:        showEndsFlag,
		Number:          numberFlag,
		SqueezeBlank:    squeezeBlankFlag,
		ShowTabs:        showTabsFlag,
		ShowNonPrinting: showNonPrintingFlag,
		Unbuffered:      posixCompatFlag,
	}
	if showAllFlag {
		opts.ShowNonPrinting = true
		opts.ShowEnds = true
		opts.ShowTabs = true
	}
	if showNonPrintingAndShowEndsFlag {
		opts.ShowNonPrinting = true
		opts.ShowEnds = true
	}
	if showNonPrintingAndShowTabsFlag {
		opts.ShowNonPrinting = true
		opts.ShowTabs = true
	}

	return opts, opts.NeedsFormatting()
}

func processSuperFlags() bool {
	_, useFlags := parseOptions()
	return useFlags
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

		opts, _ := parseOptions()
		pipeline := buildFormatPipeline(opts)
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
