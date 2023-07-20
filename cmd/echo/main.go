package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
)

var echoCmd = &cobra.Command{
	Use:   "echo [flags]… [strings]…",
	Short: "Echo the string(s) to standard output.",
	Run: func(cmd *cobra.Command, args []string) {
		showVersion, err := cmd.Flags().GetBool("version")
		if err == nil && showVersion {
			printVersion()
			return
		}

		result := strings.Join(args, " ")

		if enableStringEscape, err := cmd.Flags().GetBool("enable-escape"); err == nil && enableStringEscape {
			str, err := strconv.Unquote(`"` + result + `"`)
			if err == nil {
				result = str
			}
		}
		if newline, err := cmd.Flags().GetBool("newline"); err == nil && !newline {
			fmt.Printf("%s\n", result)
		} else {
			fmt.Printf("%s", result)
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
	fmt.Printf("version 1.0")
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
