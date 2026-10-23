package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Mode int
type Action int

const (
	LogicalMode Mode = iota
	PhysicalMode
)

const (
	ActionRun Action = iota
	ActionHelp
	ActionVersion
)

type Options struct {
	Mode Mode
}

type ParseResult struct {
	Action Action
	Mode   Mode
}

func printHelp() string {
	return "Usage:\n" +
		"pwd [OPTION]...\n" +
		"\n" +
		"Description:\n" +
		"Print the full filename of the current working directory.\n" +
		"\n" +
		"  -L, --logical\t\tuse PWD from environment, even if it contains symlinks (default)\n" +
		"  -P, --physical\tavoid all symlinks\n" +
		"  -h, --help\t\tdisplay this help and exit\n" +
		"  -v, --version\t\toutput version information and exit\n"
}

func printVersion() error {
	_, err := fmt.Println("pwd (coreutils-go) 0.1.0")
	return err
}

func parseArgs(args []string) (ParseResult, error) {
	var result ParseResult
	for _, arg := range args {
		if arg == "--" {
			break
		}

		switch arg {
		case "--help":
			result.Action = ActionHelp
			return result, nil
		case "--version":
			result.Action = ActionVersion
			return result, nil
		case "--logical":
			result.Mode = LogicalMode
			continue
		case "--physical":
			result.Mode = PhysicalMode
			continue
		}

		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			for _, opt := range arg[1:] {
				switch opt {
				case 'h':
					result.Action = ActionHelp
					return result, nil
				case 'v':
					result.Action = ActionVersion
					return result, nil
				case 'L':
					result.Mode = LogicalMode
				case 'P':
					result.Mode = PhysicalMode
				default:
					return result, fmt.Errorf("invalid option -- '%c'", opt)
				}
			}
			continue
		}
	}
	return result, nil
}

func explodePath(path string) []string {
	return strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == filepath.Separator
	})
}

func hasDotComponents(path string) bool {
	for _, segment := range explodePath(path) {
		if segment == "." || segment == ".." {
			return true
		}
	}
	return false
}

func getPhysicalDir() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}

func getLogicalDir() (string, error) {
	pwd := os.Getenv("PWD")
	if pwd != "" && filepath.IsAbs(pwd) && !hasDotComponents(pwd) {
		pwdStat, errPwd := os.Stat(pwd)
		dotStat, errDot := os.Stat(".")
		if errPwd == nil && errDot == nil && os.SameFile(pwdStat, dotStat) {
			return pwd, nil
		}
	}

	return getPhysicalDir()
}

func getWorkingDir(mode Mode) (string, error) {
	if mode == PhysicalMode {
		return getPhysicalDir()
	}
	return getLogicalDir()
}

func main() {
	result, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "pwd: %v\n", err)
		os.Exit(1)
	}
	switch result.Action {
	case ActionVersion:
		_ = printVersion()
	case ActionHelp:
		fmt.Print(printHelp())
	case ActionRun:
		path, err := getWorkingDir(result.Mode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "pwd: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(path)
	}
}
