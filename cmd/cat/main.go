package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var catCmd = &cobra.Command{
	Use:   "cat [option] [file]…",
	Short: "cat - concatenate files and print on the standard output",
	Run: func(cmd *cobra.Command, args []string) {
		for _, filename := range args {
			data, err := os.ReadFile(filename)
			if err != nil {
				panic(err)
			}
			fmt.Print(string(data))
		}
	},
}

func main() {
	if err := catCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
