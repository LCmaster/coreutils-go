package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var catCmd = &cobra.Command{
	Use:   "cat [option] [file]…",
	Short: "cat - concatenate files and print on the standard output",
	Run: func(cmd *cobra.Command, args []string) {
		for _, filename := range args {
			file, err := os.Open(filename)
			if err != nil {
				panic(err)
			}

			fileScanner := bufio.NewScanner(file)
			fileScanner.Split(bufio.ScanLines)

			lineCount := 0
			for fileScanner.Scan() {
				lineCount++
				//lines = append(lines, fileScanner.Text())
				fmt.Printf("%d\t%s\r\n", lineCount, fileScanner.Text())

			}

			file.Close()
		}
	},
}

func init() {
	// TODO: catCmd.Flags().BoolP("show-all", "A", false, "equivalent to -vET")
	// TODO: catCmd.Flags().BoolP("number-nonblank", "b", false, "number nonempty output lines, overrides -n")
	// TODO: catCmd.Flags().BoolS("esoteric-e", "e", false, "equivalent to -vE")
	// TODO: catCmd.Flags().BoolP("show-ends", "E", false, "display $ at end of each line")
	// TODO: catCmd.Flags().BoolP("number", "n", false, "number all output lines")
	// TODO: catCmd.Flags().BoolP("squeeze-blank", "s", false, "suppress repeated empty output lines")
	// TODO: catCmd.Flags().BoolS("esoteric-t", "t", false, "equivalent to -vT")
	// TODO: catCmd.Flags().BoolP("show-tabs", "T", false, "display TAB characters as ^I")
	// TODO: catCmd.Flags().BoolP("show-nonprinting", "v", false, "use ^ and M- notation, except for LFD and TAB")
	catCmd.Flags().Bool("version", false, "output version information and exit")
	catCmd.Flags().BoolP("help", "h", false, "display this help and exit")
}

func main() {
	if err := catCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
