package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var noNewline bool

var echoCmd = &cobra.Command{
	Use:   "echo [flags]… [strings]…",
	Short: "Echo the string(s) to standard output.",
	Run: func(cmd *cobra.Command, args []string) {
		result := strings.Join(args, " ")
		if !noNewline {
			result += "\n"
		}
		fmt.Printf(result)
	},
}

func init() {
	echoCmd.Flags().BoolVarP(&noNewline, "no-newline", "n", false, "do not output the trailing newline")
}

func main() {
	if err := echoCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
