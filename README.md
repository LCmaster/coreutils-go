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
| `cat` | Concatenate files and print on standard output | ⏳ Planned | — |
| `yes` | Output a string repeatedly until killed | ⏳ Planned | — |
| `true` | Return a successful exit status (0) | ⏳ Planned | — |
| `false` | Return an unsuccessful exit status (1) | ⏳ Planned | — |
| `wc` | Print newline, word, and byte counts | ⏳ Planned | — |
| `head` | Output the first part of files | ⏳ Planned | — |
| `tail` | Output the last part of files | ⏳ Planned | — |
| `touch` | Change file access and modification times | ⏳ Planned | — |
| `pwd` | Print name of current/working directory | ⏳ Planned | — |

---

## Installation

### Using `go install`

Install individual utilities directly using Go:

```sh
# Install echo
go install github.com/LCmaster/coreutils-go/cmd/echo@latest
```

### Building from Source

Clone the repository and build the binaries locally:

```sh
git clone https://github.com/LCmaster/coreutils-go.git
cd coreutils-go

# Build echo
go build -o bin/echo ./cmd/echo
```

---

## Usage Example: `echo`

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

See the [cmd/echo README](cmd/echo/README.md) for detailed option tables, supported escape codes, and examples.

---

## Project Structure

```text
coreutils-go/
├── cmd/
│   └── echo/              # echo command CLI implementation
│       ├── main.go        # Cobra CLI definitions & stream writing
│       ├── main_test.go   # Comprehensive integration & unit tests
│       └── README         # echo manual & usage guide
├── pkg/
│   └── unescape/          # Shared escape sequence interpreter
│       ├── unescape.go
│       └── unescape_test.go
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