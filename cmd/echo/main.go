package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func printHelp() {
	helpMessage := "" +
		"Echo the STRING(s) to standard output.\n" +
		"\n" +
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

	fmt.Printf(helpMessage)
}

func printVersion() {
	fmt.Printf("version 1.0")
}

func main() {
	index := 0

	newline := true
	escape := false

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			if arg == "-v" || arg == "--version" {
				printVersion()
				return
			}
			if arg == "-h" || arg == "--help" {
				printHelp()
				return
			}

			for _, flag := range arg[1:] {
				switch flag {
				case 'n':
					escape = false
				case 'E':
					escape = false
				case 'e':
					escape = true
				}
			}

		} else {
			index = i
			break
		}
	}

	stringToPrint := strings.Join(args[index:], " ")

	if !escape {
		fmt.Printf(stringToPrint)
	} else if newstr, err := strconv.Unquote("\"" + stringToPrint + "\""); err == nil {
		fmt.Printf(newstr)
	}
	if newline {
		fmt.Println()
	}
}
