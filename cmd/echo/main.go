package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/LCmaster/coreutils-go/pkg/unescape"
	"github.com/spf13/cobra"
)

func resolveEscapeFlag(cmd *cobra.Command, rawArgs []string) bool {
	enableFlag := cmd.Flags().Lookup("enable-escape")
	disableFlag := cmd.Flags().Lookup("disable-escape")

	enableChanged := enableFlag != nil && enableFlag.Changed
	disableChanged := disableFlag != nil && disableFlag.Changed

	if enableChanged && !disableChanged {
		return true
	}
	if !enableChanged && disableChanged {
		return false
	}
	if enableChanged && disableChanged {
		lastEscape := false
		for _, arg := range rawArgs {
			if arg == "--enable-escape" {
				lastEscape = true
			} else if arg == "--disable-escape" {
				lastEscape = false
			} else if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && len(arg) > 1 {
				for _, r := range arg[1:] {
					if r == 'e' {
						lastEscape = true
					} else if r == 'E' {
						lastEscape = false
					}
				}
			}
		}
		return lastEscape
	}

	return false
}

var echoCmd = &cobra.Command{
	Use:   "echo [flags]… [strings]…",
	Short: "Echo the string(s) to standard output.",
	Run: func(cmd *cobra.Command, args []string) {
		showVersion, err := cmd.Flags().GetBool("version")
		if err == nil && showVersion {
			printVersion()
			return
		}

		enableEscape := resolveEscapeFlag(cmd, os.Args[1:])
		newline, _ := cmd.Flags().GetBool("newline")

		writer := bufio.NewWriter(cmd.OutOrStdout())
		defer writer.Flush()

		halt := false
		for i, arg := range args {
			if i > 0 {
				if err := writer.WriteByte(' '); err != nil {
					return
				}
			}

			if enableEscape {
				var out string
				out, halt = unescape.Unescape(arg)
				if _, err := writer.WriteString(out); err != nil {
					return
				}
				if halt {
					break
				}
			} else {
				if _, err := writer.WriteString(arg); err != nil {
					return
				}
			}
		}

		if !newline && !halt {
			_ = writer.WriteByte('\n')
		}
	},
}

func printHelp() string {
	helpMessage := "" +
		"Usage:\n" +
		"echo [OPTIONS]... [STRING]...\n" +
		"\n" +
		"Description:\n" +
		"echo writes each given string to standard output, with a space between each and a newline after the last one.\n" +
		"  -n\t\t\tdo not output the trailing newline\n" +
		"  -e\t\t\tenable interpretation of backslash escapes\n" +
		"  -E\t\t\tdisable interpretation of backslash escapes (default)\n" +
		"  -h, --help\tdisplay this help and exit\n" +
		"  -v, --version\toutput version information and exit\n" +
		"\n" +
		"If -e is in effect, the following sequences are recognized:\n" +
		"  \\\\\t\tbackslash\n" +
		"  \\a\t\talert (BEL)\n" +
		"  \\b\t\tbackspace\n" +
		"  \\c\t\tproduce no further output\n" +
		"  \\e\t\tescape\n" +
		"  \\f\t\tform feed\n" +
		"  \\n\t\tnew line\n" +
		"  \\r\t\tcarriage return\n" +
		"  \\t\t\thorizontal tab\n" +
		"  \\v\t\tvertical tab\n" +
		"  \\0NNN\t\tbyte with octal value NNN (1 to 3 digits)\n" +
		"  \\xHH\t\tbyte with hexadecimal value HH (1 to 2 digits)\n"

	return helpMessage
}

func printVersion() {
	fmt.Println("version 0.1.0")
}

func init() {
	echoCmd.Flags().BoolS("newline", "n", false, "do not output the trailing newline")
	echoCmd.Flags().BoolS("disable-escape", "E", true, "disable interpretation of backslash escapes (default)")
	echoCmd.Flags().BoolS("enable-escape", "e", false, "enable interpretation of backslash escapes")
	echoCmd.Flags().BoolP("version", "v", false, "output version information and exit")
	echoCmd.Flags().BoolP("help", "h", false, "display this help and exit")

	echoCmd.SetUsageTemplate(printHelp())
}

func main() {
	if err := echoCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
