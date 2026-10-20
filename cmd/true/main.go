package main

import (
	"fmt"
	"os"
)

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

func printVersion() error {
	_, err := fmt.Println("true (coreutils-go) 0.1.0")
	return err
}

func main() {
	// GNU true specification:
	// --help and --version are only recognized when they are the sole argument.
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "-h", "--help":
			if _, err := fmt.Print(printHelp()); err != nil {
				os.Exit(1)
			}
			return
		case "-v", "--version":
			if err := printVersion(); err != nil {
				os.Exit(1)
			}
			return
		}
	}

	// In all other cases, exit 0 successfully
}
