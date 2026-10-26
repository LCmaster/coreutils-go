# coreutils-go

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](go.mod)

A modern, fast, cross-platform rewrite of the classic **GNU Coreutils** in pure Go, with zero external runtime dependencies.

---

## Overview

`coreutils-go` aims to provide clean, idiomatic, and highly compatible replacements for essential Unix/Linux core utilities. Every utility is designed to run consistently across **Linux**, **macOS**, and **Windows**.

### Key Design Goals

- **GNU / POSIX Compatibility**: Strict adherence to standard flag behaviors, exit codes, and output formats.
- **High Performance**: Optimized buffered I/O (`bufio`) and zero-allocation parsing routines.
- **Cross-Platform**: Native execution on Windows, Linux, and macOS without requiring a POSIX emulation layer or WSL.
- **Modular Architecture**: Shared libraries located in `pkg/` for reusable components (escape sequence parsing, stream handling, CLI formatting).
- **Comprehensive Testing**: Full test suites featuring both unit tests and end-to-end integration tests.

---

## Implemented Utilities

| Utility | Description | Status | Documentation |
| :--- | :--- | :---: | :--- |
| [`echo`](cmd/echo) | Write arguments to standard output | ✅ Stable | [README](cmd/echo/README.md) |
| [`cat`](cmd/cat) | Concatenate files and print on standard output | ✅ Stable | [README](cmd/cat/README.md) |
| [`pwd`](cmd/pwd) | Print name of current/working directory | ✅ Stable | [README](cmd/pwd/README.md) |
| [`yes`](cmd/yes) | Output a string repeatedly until killed | ✅ Stable | [README](cmd/yes/README.md) |
| [`true`](cmd/true) | Return a successful exit status (0) | ✅ Stable | [README](cmd/true/README.md) |
| [`false`](cmd/false) | Return an unsuccessful exit status (1) | ✅ Stable | [README](cmd/false/README.md) |
| [`head`](cmd/head) | Output the first part of files | ✅ Stable | [README](cmd/head/README.md) |
| `wc` | Print newline, word, and byte counts | ⏳ Planned | — |
| `tail` | Output the last part of files | ⏳ Planned | — |
| `touch` | Change file access and modification times | ⏳ Planned | — |

---

## Installation

### Using `go install`

Install utilities directly using Go:

```sh
# Install specific utilities
go install github.com/LCmaster/coreutils-go/cmd/echo@latest
go install github.com/LCmaster/coreutils-go/cmd/cat@latest
go install github.com/LCmaster/coreutils-go/cmd/pwd@latest
go install github.com/LCmaster/coreutils-go/cmd/yes@latest
go install github.com/LCmaster/coreutils-go/cmd/true@latest
go install github.com/LCmaster/coreutils-go/cmd/false@latest
go install github.com/LCmaster/coreutils-go/cmd/head@latest

# Or install all utilities
go install github.com/LCmaster/coreutils-go/cmd/...@latest
```

### Building from Source

Clone the repository and build the binaries locally:

```sh
git clone https://github.com/LCmaster/coreutils-go.git
cd coreutils-go

# Build all utilities
go build -o bin/ ./cmd/...
```

---

## Usage Examples

### `echo`

```sh
# Basic text printing
$ echo Hello, World!
Hello, World!

# Suppress trailing newline
$ echo -n "Enter password: "

# Interpret escape sequences (tabs, newlines, colors)
$ echo -e "Line 1\nLine 2\tIndented"
$ echo -e "\e[32mSuccess:\e[0m Task completed."

# Early termination with \c
$ echo -e "First part\cThis will not be printed"
First part$
```

See [cmd/echo/README.md](cmd/echo/README.md) for full details.

### `cat`

```sh
# Concatenate files to standard output
$ cat file1.txt file2.txt

# Read from standard input
$ cat -

# Number all output lines
$ cat -n file.txt
     1	First line
     2	Second line

# Number non-empty lines only
$ cat -b file.txt

# Squeeze repeated empty lines
$ cat -s file.txt

# Display $ at end of each line and display tabs as ^I
$ cat -ET file.txt

# Show all non-printing characters
$ cat -A file.txt
```

See [cmd/cat/README.md](cmd/cat/README.md) for full details.

### `pwd`

```sh
# Print current working directory (logical by default)
$ pwd
/home/user/workspace

# Print physical directory resolving all symlinks
$ pwd -P
/actual/storage/mount/workspace
```

See [cmd/pwd/README.md](cmd/pwd/README.md) for full details.

### `yes`

```sh
# Default repetition ('y')
$ yes
y
y
...

# Custom string repetition
$ yes custom text
custom text
custom text
...

# Automated confirmations
$ yes | rm -i *.tmp
```

See [cmd/yes/README.md](cmd/yes/README.md) for full details.

### `true`

```sh
# Exit successfully (code 0)
$ true
$ echo $?
0

# Any arbitrary arguments or flags are ignored
$ true ignored arguments --even-flags
$ echo $?
0
```

See [cmd/true/README.md](cmd/true/README.md) for full details.

### `false`

```sh
# Exit unsuccessfully (code 1)
$ false
$ echo $?
1

# Any arbitrary arguments or flags are ignored
$ false ignored arguments --even-flags
$ echo $?
1
```

See [cmd/false/README.md](cmd/false/README.md) for full details.

### `head`

```sh
# Output the first 10 lines (default)
$ head file.txt
$ head -

# Custom line count
$ head -n 5 file.txt
$ head -5 file.txt

# Print all but the trailing 3 lines
$ head -n -3 file.txt

# Output by byte count (with optional multiplier suffixes)
$ head -c 64 file.txt
$ head -c 1K file.txt

# Print all bytes except the trailing 10 bytes
$ head -c -10 file.txt

# Multiple files with headers
$ head file1.txt file2.txt
==> file1.txt <==
...

==> file2.txt <==
...

# Suppress headers or force header on single file / stdin
$ head -q file1.txt file2.txt
$ head -v file.txt
$ echo "hello" | head -v
==> standard input <==
hello

# NUL-delimited records
$ head -z -n 2 records.txt
```

See [cmd/head/README.md](cmd/head/README.md) for full details.

---

## Project Structure

```text
coreutils-go/
├── cmd/
│   ├── cat/               # cat command CLI implementation
│   │   ├── cat.go         # Core stream/file processing, cycle detection & safety
│   │   ├── transform.go   # Functional line transformer pipeline & combinators
│   │   ├── main.go        # CLI parser & streaming lifecycle
│   │   ├── main_test.go   # Complete integration & unit test suite
│   │   └── README.md      # cat manual & usage guide
│   ├── echo/              # echo command CLI implementation
│   │   ├── main.go        # Arguments & escape sequence writer
│   │   ├── main_test.go   # Comprehensive integration & unit tests
│   │   └── README.md      # echo manual & usage guide
│   ├── false/             # false command CLI implementation
│   │   ├── main.go        # false command entry point
│   │   ├── main_test.go   # Complete test suite
│   │   └── README.md      # false manual & usage guide
│   ├── head/              # head command CLI implementation
│   │   ├── head.go        # Positive/negative lines & bytes filtering engines
│   │   ├── main.go        # Multi-file loop, headers, & CLI entry point
│   │   ├── main_test.go   # Complete integration & unit test suite
│   │   └── README.md      # head manual & usage guide
│   ├── pwd/               # pwd command CLI implementation
│   │   ├── predicate.go   # Functional combinators & path validation predicates
│   │   ├── pwd.go         # Domain model & path resolution strategies
│   │   ├── main.go        # CLI argument parsing, streaming I/O & lifecycle
│   │   ├── main_test.go   # Comprehensive integration & unit tests
│   │   └── README.md      # pwd manual & usage guide
│   ├── true/              # true command CLI implementation
│   │   ├── main.go        # true command entry point
│   │   ├── main_test.go   # Complete test suite
│   │   └── README.md      # true manual & usage guide
│   └── yes/               # yes command CLI implementation
│       ├── main.go        # High-performance 64KB chunk stream
│       ├── main_test.go   # Tests & throughput benchmark
│       └── README.md      # yes manual & usage guide
├── pkg/
│   ├── cli/               # Lightweight, zero-dependency GNU/POSIX argument parser
│   ├── fn/                # Functional combinators & stream pipeline helpers
│   ├── fsutils/           # Cross-platform filesystem & path introspection
│   ├── ioutils/           # Resilient I/O routines & broken pipe detection
│   ├── strutils/          # GNU string utilities & escape sequence parser
│   └── version/           # Centralized semantic version formatting
├── go.mod                 # Go module definition (Go 1.26)
├── go.sum
├── LICENSE                # MIT License
└── README.md              # Project documentation
```

---

## Testing & Verification

Run the entire test suite across all utilities and packages:

```sh
# Run all tests
go test -v ./...

# Run tests with race detection
go test -v -race ./...

# Test a specific utility
go test -v ./cmd/echo
```

---

## Contributing

Contributions are welcome! If you would like to implement a new utility or enhance an existing one:

1. Fork the repository.
2. Create your feature branch (`git checkout -b feat/add-cat`).
3. Ensure all code is formatted (`go fmt ./...`) and covered by tests.
4. Verify tests pass (`go test ./...`).
5. Open a Pull Request.

---

## License

This project is licensed under the terms of the [MIT License](LICENSE).