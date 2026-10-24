// Package cli provides a lightweight, zero-dependency POSIX and GNU compliant
// command-line option parser for the coreutils-go suite.
package cli

import (
	"fmt"
	"strings"
)

// Action represents the outcome of command line argument parsing.
type Action int

const (
	// ActionRun indicates normal execution with parsed options and operands.
	ActionRun Action = iota
	// ActionHelp indicates --help was requested.
	ActionHelp
	// ActionVersion indicates --version was requested.
	ActionVersion
)

// Flag represents a command-line flag specification.
type Flag struct {
	Short       rune
	Long        string
	Description string
	TakesArg    bool
	Handler     func(val string) error
}

// Parser parses command line arguments according to POSIX and GNU coreutils conventions.
type Parser struct {
	Name             string
	Usage            string
	flags            []*Flag
	shortMap         map[rune]*Flag
	longMap          map[string]*Flag
	numericShorthand func(num string) error
	numericErrFormat string
	defaultOperands  []string
	hasAutoHelp      bool
	hasAutoShortHelp bool
	hasAutoVersion   bool
}

// NewParser creates a new CLI parser initialized with standard GNU defaults.
func NewParser(name, usage string) *Parser {
	return &Parser{
		Name:             name,
		Usage:            usage,
		shortMap:         make(map[rune]*Flag),
		longMap:          make(map[string]*Flag),
		hasAutoHelp:      true,
		hasAutoShortHelp: true,
		hasAutoVersion:   true,
		numericErrFormat: "%s: invalid number: '%s'",
	}
}

// DisableAutoShortHelp disables treating -h as an automatic help option,
// while preserving --help (e.g. for GNU cat).
func (p *Parser) DisableAutoShortHelp() {
	p.hasAutoShortHelp = false
}

// BoolVar registers a boolean flag. If the flag is supplied, *ptr is set to true.
func (p *Parser) BoolVar(ptr *bool, short rune, long, description string) {
	p.FlagFunc(short, long, description, false, func(_ string) error {
		*ptr = true
		return nil
	})
}

// StringVar registers a string option that requires an argument.
func (p *Parser) StringVar(ptr *string, short rune, long, description string) {
	p.FlagFunc(short, long, description, true, func(val string) error {
		*ptr = val
		return nil
	})
}

// FlagFunc registers a custom flag with an evaluation callback.
func (p *Parser) FlagFunc(short rune, long, description string, takesArg bool, handler func(val string) error) {
	f := &Flag{
		Short:       short,
		Long:        long,
		Description: description,
		TakesArg:    takesArg,
		Handler:     handler,
	}
	p.flags = append(p.flags, f)
	if short != 0 {
		p.shortMap[short] = f
	}
	if long != "" {
		p.longMap[long] = f
	}
}

// EnableNumericShorthand enables parsing of legacy numeric options (e.g. -5, -20)
// commonly supported by utilities like head and tail.
func (p *Parser) EnableNumericShorthand(errFormat string, handler func(num string) error) {
	p.numericShorthand = handler
	if errFormat != "" {
		p.numericErrFormat = errFormat
	}
}

// SetDefaultOperands sets default operands to use when no positional arguments are given
// (e.g. []string{"-"} for stdin-default tools).
func (p *Parser) SetDefaultOperands(operands ...string) {
	p.defaultOperands = operands
}

func isNumeric(s string) bool {
	if !strings.HasPrefix(s, "-") || len(s) < 2 {
		return false
	}
	for _, r := range s[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Parse processes command-line arguments according to GNU/POSIX rules:
//   - Supports short option clustering (e.g. -qv -> -q, -v).
//   - Supports attached and separate option values (-n5, -n 5, --lines=5, --lines 5).
//   - Preserves negative numbers passed to options taking values (-n -5, -n-5).
//   - Stops option processing upon encountering "--".
//   - Treats single "-" as a regular positional operand (representing standard input).
//   - Formats error messages matching GNU coreutils diagnostics.
func (p *Parser) Parse(args []string) ([]string, Action, error) {
	var operands []string
	var endOfOptions bool

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if endOfOptions {
			operands = append(operands, arg)
			continue
		}

		if arg == "--" {
			endOfOptions = true
			continue
		}

		if arg == "-" || !strings.HasPrefix(arg, "-") {
			operands = append(operands, arg)
			continue
		}

		// Legacy numeric option (e.g. -5, -20)
		if p.numericShorthand != nil && isNumeric(arg) {
			if err := p.numericShorthand(arg[1:]); err != nil {
				return nil, ActionRun, fmt.Errorf(p.numericErrFormat, p.Name, arg[1:])
			}
			continue
		}

		// Long option (--help, --lines=10, --bytes 20)
		if strings.HasPrefix(arg, "--") {
			name := arg[2:]
			var val string
			hasVal := false

			if eqIdx := strings.IndexByte(name, '='); eqIdx != -1 {
				val = name[eqIdx+1:]
				name = name[:eqIdx]
				hasVal = true
			}

			if name == "help" && p.hasAutoHelp && p.longMap["help"] == nil {
				return nil, ActionHelp, nil
			}
			if name == "version" && p.hasAutoVersion && p.longMap["version"] == nil {
				return nil, ActionVersion, nil
			}

			flag, exists := p.longMap[name]
			if !exists {
				return nil, ActionRun, fmt.Errorf("%s: unrecognized option '--%s'", p.Name, name)
			}

			if flag.TakesArg {
				if !hasVal {
					if i+1 >= len(args) {
						return nil, ActionRun, fmt.Errorf("%s: option '--%s' requires an argument", p.Name, name)
					}
					i++
					val = args[i]
				}
				if err := flag.Handler(val); err != nil {
					return nil, ActionRun, err
				}
			} else {
				if hasVal {
					return nil, ActionRun, fmt.Errorf("%s: option '--%s' doesn't allow an argument", p.Name, name)
				}
				if err := flag.Handler(""); err != nil {
					return nil, ActionRun, err
				}
			}
			continue
		}

		// Short option cluster (-x, -qv, -n10, -n-5)
		runes := []rune(arg[1:])
		for j := 0; j < len(runes); j++ {
			r := runes[j]

			// Registered flags take priority over auto help/version
			flag, exists := p.shortMap[r]
			if !exists {
				if r == 'h' && p.hasAutoHelp && p.hasAutoShortHelp {
					return nil, ActionHelp, nil
				}
				return nil, ActionRun, fmt.Errorf("%s: invalid option -- '%c'", p.Name, r)
			}

			if flag.TakesArg {
				var val string
				if j+1 < len(runes) {
					// Attached value: e.g. -n10 or -n-5
					val = string(runes[j+1:])
				} else {
					// Separate argument: e.g. -n 10
					if i+1 >= len(args) {
						return nil, ActionRun, fmt.Errorf("%s: option requires an argument -- '%c'", p.Name, r)
					}
					i++
					val = args[i]
				}
				if err := flag.Handler(val); err != nil {
					return nil, ActionRun, err
				}
				break
			} else {
				if err := flag.Handler(""); err != nil {
					return nil, ActionRun, err
				}
			}
		}
	}

	if len(operands) == 0 && len(p.defaultOperands) > 0 {
		operands = append(operands, p.defaultOperands...)
	}

	return operands, ActionRun, nil
}
