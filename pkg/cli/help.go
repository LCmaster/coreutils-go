package cli

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// HelpOption defines a command-line option for formatted help output.
type HelpOption struct {
	Flags       string // e.g. "-n, --number", "-c, --bytes=[-]NUM", "      --help"
	Description string // option description (supports multi-line text)
}

// Opt creates a HelpOption with standard GNU flag formatting and alignment.
func Opt(short rune, long, arg, description string) HelpOption {
	var flagStr string
	longs := splitLongNames(long)

	switch {
	case short != 0 && len(longs) > 0:
		var longParts []string
		for _, l := range longs {
			if arg != "" {
				longParts = append(longParts, fmt.Sprintf("--%s=%s", l, arg))
			} else {
				longParts = append(longParts, fmt.Sprintf("--%s", l))
			}
		}
		flagStr = fmt.Sprintf("-%c, %s", short, strings.Join(longParts, ", "))
	case short == 0 && len(longs) > 0:
		var longParts []string
		for _, l := range longs {
			if arg != "" {
				longParts = append(longParts, fmt.Sprintf("--%s=%s", l, arg))
			} else {
				longParts = append(longParts, fmt.Sprintf("--%s", l))
			}
		}
		flagStr = fmt.Sprintf("      %s", strings.Join(longParts, ", "))
	case short != 0 && len(longs) == 0:
		if arg != "" {
			flagStr = fmt.Sprintf("-%c %s", short, arg)
		} else {
			flagStr = fmt.Sprintf("-%c", short)
		}
	}

	return HelpOption{
		Flags:       flagStr,
		Description: description,
	}
}

// HelpConfig contains all documentation metadata needed to generate a standardized GNU help manual.
type HelpConfig struct {
	Usage       []string     // Synopses: e.g. ["cat [OPTION]... [FILE]..."]
	Description string       // Utility description / summary
	Options     []HelpOption // Registered command-line options
	Notes       string       // Optional extra notes / format descriptions
	Examples    []string     // Optional example command lines
}

// FormatHelp renders a standardized, POSIX/GNU-compliant help manual.
func FormatHelp(cfg HelpConfig) string {
	var b strings.Builder

	// 1. Usage section
	if len(cfg.Usage) > 0 {
		b.WriteString("Usage: ")
		b.WriteString(cfg.Usage[0])
		b.WriteString("\n")
		for _, alt := range cfg.Usage[1:] {
			b.WriteString("  or:  ")
			b.WriteString(alt)
			b.WriteString("\n")
		}
	}

	// 2. Description section
	if cfg.Description != "" {
		b.WriteString("\nDescription:\n")
		b.WriteString(cfg.Description)
		b.WriteString("\n")
	}

	// 3. Options section
	if len(cfg.Options) > 0 {
		b.WriteString("\n")
		flagColWidth := 24
		for _, opt := range cfg.Options {
			w := utf8.RuneCountInString(opt.Flags)
			if w > flagColWidth && w <= 30 {
				flagColWidth = w
			}
		}

		for _, opt := range cfg.Options {
			b.WriteString("  ")
			b.WriteString(opt.Flags)
			descLines := strings.Split(opt.Description, "\n")
			if len(descLines) == 0 || opt.Description == "" {
				b.WriteString("\n")
				continue
			}

			flagWidth := utf8.RuneCountInString(opt.Flags)
			padding := max((flagColWidth+2)-flagWidth, 2)
			b.WriteString(strings.Repeat(" ", padding))
			b.WriteString(descLines[0])
			b.WriteString("\n")

			contIndent := strings.Repeat(" ", 2+flagColWidth+2)
			for _, line := range descLines[1:] {
				b.WriteString(contIndent)
				b.WriteString(line)
				b.WriteString("\n")
			}
		}
	}

	// 4. Notes / Supplementary format section
	if cfg.Notes != "" {
		b.WriteString("\n")
		b.WriteString(cfg.Notes)
		b.WriteString("\n")
	}

	// 5. Examples section
	if len(cfg.Examples) > 0 {
		b.WriteString("\nExamples:\n")
		for _, ex := range cfg.Examples {
			b.WriteString("  ")
			b.WriteString(ex)
			b.WriteString("\n")
		}
	}

	return b.String()
}

// SetUsage sets or overrides the usage synopsis lines for the command.
func (c *Command) SetUsage(usage ...string) {
	c.Helper.Usage = usage
	for _, u := range usage {
		if strings.Contains(u, "[ignored command line arguments]") {
			c.Config.SoleArgumentHelpVersion = true
		}
	}
}

// SetDescription sets the command description for the help manual.
func (c *Command) SetDescription(desc string) {
	c.Helper.Description = desc
}

// SetNotes sets extra syntax details or format descriptions for the help manual.
func (c *Command) SetNotes(notes string) {
	c.Helper.Notes = notes
}

// AddExample adds an example usage line to the help manual.
func (c *Command) AddExample(example string) {
	c.Helper.Examples = append(c.Helper.Examples, example)
}

// AddHelpOption adds a manual / synthetic option to the help manual.
func (c *Command) AddHelpOption(opt HelpOption) {
	c.Helper.Options = append(c.Helper.Options, opt)
}

// Help formats the complete GNU help manual using registered flags and documentation metadata.
func (c *Command) Help() string {
	options := append(c.Flags.ToHelpOptions(), c.Helper.Options...)

	hasHelp := false
	hasVersion := false
	for _, opt := range options {
		if strings.Contains(opt.Flags, "--help") {
			hasHelp = true
		}
		if strings.Contains(opt.Flags, "--version") {
			hasVersion = true
		}
	}

	if c.Config.AutoHelp && !hasHelp {
		short := rune(0)
		if c.Config.AutoShortHelp {
			short = 'h'
		}
		options = append(options, Opt(short, "help", "", "display this help and exit"))
	}
	if c.Config.AutoVersion && !hasVersion {
		short := rune(0)
		if c.Config.AutoShortVersion {
			short = 'v'
		}
		options = append(options, Opt(short, "version", "", "output version information and exit"))
	}

	return FormatHelp(HelpConfig{
		Usage:       c.Helper.Usage,
		Description: c.Helper.Description,
		Options:     options,
		Notes:       c.Helper.Notes,
		Examples:    c.Helper.Examples,
	})
}
