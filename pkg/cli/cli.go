// Package cli provides a lightweight, zero-dependency POSIX and GNU compliant
// command-line argument parser for the coreutils-go suite.
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

// Parser is a pure argument token parsing engine.
// It parses an argument list against a FlagSet and Config without any knowledge
// of documentation, usage synopses, or help presentation.
type Parser struct {
	name             string // used for diagnostics matching GNU format: "<name>: invalid option -- 'x'"
	flags            *FlagSet
	config           Config
	operands         []string
	explicitOperands []string
	action           Action
}

// NewParser creates a new pure argument parser instance.
func NewParser(name string, flags *FlagSet, config Config) *Parser {
	if flags == nil {
		flags = NewFlagSet()
	}
	return &Parser{
		name:   name,
		flags:  flags,
		config: config,
		action: ActionRun,
	}
}

// Operands returns positional operands collected during Parse
// (including DefaultOperands if no explicit operands were provided).
func (p *Parser) Operands() []string {
	return p.operands
}

// ExplicitOperands returns only the positional operands explicitly passed on the
// command line, excluding any fallback DefaultOperands.
func (p *Parser) ExplicitOperands() []string {
	return p.explicitOperands
}

// UsedDefaultOperands reports whether DefaultOperands were substituted because no
// explicit positional operands were supplied.
func (p *Parser) UsedDefaultOperands() bool {
	return len(p.explicitOperands) == 0 && len(p.operands) > 0
}

// Args is an alias for Operands matching Go standard library conventions.
func (p *Parser) Args() []string {
	return p.operands
}

// Action returns the action determined by Parse (ActionRun, ActionHelp, or ActionVersion).
func (p *Parser) Action() Action {
	return p.action
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
func (p *Parser) Parse(args []string) error {
	p.operands = nil
	p.explicitOperands = nil
	p.action = ActionRun

	if p.config.SoleArgumentHelpVersion {
		if len(args) == 1 {
			switch args[0] {
			case "--help":
				if p.config.AutoHelp {
					p.action = ActionHelp
					return nil
				}
			case "-h":
				if p.config.AutoHelp && p.config.AutoShortHelp {
					p.action = ActionHelp
					return nil
				}
			case "--version":
				if p.config.AutoVersion {
					p.action = ActionVersion
					return nil
				}
			case "-v":
				if p.config.AutoVersion && p.config.AutoShortVersion {
					p.action = ActionVersion
					return nil
				}
			}
		}
		p.explicitOperands = args
		p.operands = args
		return nil
	}

	var operands []string
	var endOfOptions bool

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if endOfOptions {
			operands = append(operands, arg)
			continue
		}

		if arg == "--" {
			if p.config.StopOnUnknown {
				operands = append(operands, args[i:]...)
				break
			}
			endOfOptions = true
			continue
		}

		if arg == "-" || !strings.HasPrefix(arg, "-") {
			if p.config.StopOnUnknown {
				operands = append(operands, args[i:]...)
				break
			}
			operands = append(operands, arg)
			continue
		}

		// Legacy numeric option (e.g. -5, -20)
		if p.config.NumericShorthand != nil && isNumeric(arg) {
			if err := p.config.NumericShorthand(arg[1:]); err != nil {
				return fmt.Errorf(p.config.NumericErrFormat, p.name, arg[1:])
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

			flag, exists := p.flags.LookupLong(name)

			if name == "help" && p.config.AutoHelp && !exists {
				if p.config.StopOnUnknown && i > 0 {
					operands = append(operands, args[i:]...)
					break
				}
				p.action = ActionHelp
				return nil
			}
			if name == "version" && p.config.AutoVersion && !exists {
				if p.config.StopOnUnknown && i > 0 {
					operands = append(operands, args[i:]...)
					break
				}
				p.action = ActionVersion
				return nil
			}

			if !exists {
				if p.config.StopOnUnknown {
					operands = append(operands, args[i:]...)
					break
				}
				return fmt.Errorf("%s: unrecognized option '--%s'", p.name, name)
			}

			if flag.TakesArg {
				if !hasVal {
					if i+1 >= len(args) {
						return fmt.Errorf("%s: option '--%s' requires an argument", p.name, name)
					}
					i++
					val = args[i]
				}
				if err := flag.Handler(val); err != nil {
					return err
				}
			} else {
				if hasVal {
					return fmt.Errorf("%s: option '--%s' doesn't allow an argument", p.name, name)
				}
				if err := flag.Handler(""); err != nil {
					return err
				}
			}
			continue
		}

		// Short option cluster (-x, -qv, -n10, -n-5)
		runes := []rune(arg[1:])

		if p.config.StopOnUnknown {
			allKnown := true
			for _, r := range runes {
				if _, ok := p.flags.LookupShort(r); !ok {
					if !(r == 'h' && p.config.AutoHelp && p.config.AutoShortHelp) &&
						!(r == 'v' && p.config.AutoVersion && p.config.AutoShortVersion) {
						allKnown = false
						break
					}
				}
			}
			if !allKnown {
				operands = append(operands, args[i:]...)
				break
			}
		}

		for j := 0; j < len(runes); j++ {
			r := runes[j]

			// Registered flags take priority over auto help/version
			flag, exists := p.flags.LookupShort(r)
			if !exists {
				if r == 'h' && p.config.AutoHelp && p.config.AutoShortHelp {
					if p.config.StopOnUnknown && i > 0 {
						operands = append(operands, args[i:]...)
						break
					}
					p.action = ActionHelp
					return nil
				}
				if r == 'v' && p.config.AutoVersion && p.config.AutoShortVersion {
					if p.config.StopOnUnknown && i > 0 {
						operands = append(operands, args[i:]...)
						break
					}
					p.action = ActionVersion
					return nil
				}
				if p.config.StopOnUnknown {
					operands = append(operands, args[i:]...)
					break
				}
				return fmt.Errorf("%s: invalid option -- '%c'", p.name, r)
			}

			if flag.TakesArg {
				var val string
				if j+1 < len(runes) {
					val = string(runes[j+1:])
				} else {
					if i+1 >= len(args) {
						return fmt.Errorf("%s: option requires an argument -- '%c'", p.name, r)
					}
					i++
					val = args[i]
				}
				if err := flag.Handler(val); err != nil {
					return err
				}
				break
			} else {
				if err := flag.Handler(""); err != nil {
					return err
				}
			}
		}
	}

	p.explicitOperands = operands
	if len(operands) == 0 && len(p.config.DefaultOperands) > 0 {
		operands = append([]string(nil), p.config.DefaultOperands...)
	}

	p.operands = operands
	return nil
}
