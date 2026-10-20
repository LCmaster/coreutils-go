package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

const bufferSize = 64 * 1024 // 64 KB memory buffer for optimal I/O throughput

func printHelp() string {
	return "Usage:\n" +
		"yes [STRING]...\n" +
		"  or:  yes OPTION\n" +
		"\n" +
		"Description:\n" +
		"Repeatedly output a line with all specified STRING(s), or 'y'.\n" +
		"\n" +
		"  -h, --help     display this help and exit\n" +
		"  -v, --version  output version information and exit\n"
}

func printVersion() error {
	_, err := fmt.Println("yes (coreutils-go) 0.1.0")
	return err
}

// buildBuffer creates a reusable chunk buffer pre-filled with line repetitions
// up to bufferSize to minimize system call overhead.
func buildBuffer(line []byte) []byte {
	if len(line) >= bufferSize {
		return line
	}

	repeats := bufferSize / len(line)
	if repeats == 0 {
		repeats = 1
	}

	return bytes.Repeat(line, repeats)
}

// runYes continuously writes the given arguments or 'y' to w until an error occurs.
func runYes(w io.Writer, args []string) error {
	var line []byte
	if len(args) == 0 {
		line = []byte("y\n")
	} else {
		line = []byte(strings.Join(args, " ") + "\n")
	}

	buf := buildBuffer(line)
	for {
		if _, err := w.Write(buf); err != nil {
			return err
		}
	}
}

func main() {
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

	_ = runYes(os.Stdout, os.Args[1:])
}
