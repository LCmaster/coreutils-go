package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var trueCmd = &cobra.Command{
	Use:   "true [ignored command line arguments]",
	Short: "Do nothing, successfully.",
	Run: func(cmd *cobra.Command, args []string) {
		// Nothing to be done here...
	},
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
}

func printHelp() string {
	return "Usage:\n" +
		"true [ignored command line arguments]\n" +
		"  or:  true OPTION\n" +
		"\n" +
		"Description:\n" +
		"Exit with a status code indicating success.\n" +
		"\n" +
		"  -h, --help     display this help and exit\n" +
		"  -v, --version  output version information and exit\n"
}

func printVersion() {
	fmt.Println("version 0.1.0")
}

func init() {
	trueCmd.Flags().BoolP("version", "v", false, "output version information and exit")
	trueCmd.Flags().BoolP("help", "h", false, "display this help and exit")

	trueCmd.SetUsageTemplate(printHelp())
}

func main() {
	// GNU true specification:
	// --help and --version are only recognized when they are the sole argument.
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "-h", "--help":
			fmt.Print(printHelp())
			return
		case "-v", "--version":
			printVersion()
			return
		}
	}

	// In all other cases (no args, multiple args, or ignored flags), exit 0 immediately
}
