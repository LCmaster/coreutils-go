package cli

import (
	"fmt"
	"strings"
)

// HelpOption defines a command-line option for formatted help output.
type HelpOption struct {
	Flags       string // e.g. "-n, --number", "-c, --bytes=[-]NUM", "      --help"
	Description string // option description (supports multi-line text)
}

// Opt creates a HelpOption with standard GNU flag formatting and alignment.
func Opt(short rune, long, arg, description string) HelpOption {
	var flagStr string
	switch {
	case short != 0 && long != "":
		if arg != "" {
			flagStr = fmt.Sprintf("-%c, --%s=%s", short, long, arg)
		} else {
			flagStr = fmt.Sprintf("-%c, --%s", short, long)
		}
	case short == 0 && long != "":
		if arg != "" {
			flagStr = fmt.Sprintf("      --%s=%s", long, arg)
		} else {
			flagStr = fmt.Sprintf("      --%s", long)
		}
	case short != 0 && long == "":
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

// HelpConfig contains all information needed to generate a standardized GNU help page.
type HelpConfig struct {
	Usage       []string     // Synopses: e.g. ["cat [OPTION]... [FILE]..."] or ["yes [STRING]...", "yes OPTION"]
	Description string       // Utility description / summary
	Options     []HelpOption // Registered command-line options
	Notes       string       // Optional extra notes / format descriptions (e.g., escape sequences, multipliers)
	Examples    []string     // Optional example command lines
}

// FormatHelp renders a standardized, POSIX/GNU-compliant help manual.
// Output layout:
//
//	Usage: <synopsis 1>
//	  or:  <synopsis 2>
//
//	Description:
//	<description>
//
//	  <flags>    <description>
//
//	<notes>
//
//	Examples:
//	  <example 1>
func FormatHelp(cfg HelpConfig) string {
	var b strings.Builder

	// 1. Usage section
	if len(cfg.Usage) > 0 {
		b.WriteString("Usage: " + cfg.Usage[0] + "\n")
		for _, alt := range cfg.Usage[1:] {
			b.WriteString("  or:  " + alt + "\n")
		}
	}

	// 2. Description section
	if cfg.Description != "" {
		b.WriteString("\nDescription:\n")
		b.WriteString(cfg.Description + "\n")
	}

	// 3. Options section
	if len(cfg.Options) > 0 {
		b.WriteString("\n")
		flagColWidth := 24
		for _, opt := range cfg.Options {
			if len(opt.Flags) > flagColWidth && len(opt.Flags) <= 30 {
				flagColWidth = len(opt.Flags)
			}
		}

		for _, opt := range cfg.Options {
			b.WriteString("  " + opt.Flags)
			descLines := strings.Split(opt.Description, "\n")
			if len(descLines) == 0 || opt.Description == "" {
				b.WriteString("\n")
				continue
			}

			padding := (flagColWidth + 2) - len(opt.Flags)
			if padding < 2 {
				padding = 2
			}
			b.WriteString(strings.Repeat(" ", padding))
			b.WriteString(descLines[0] + "\n")

			contIndent := strings.Repeat(" ", 2+flagColWidth+2)
			for _, line := range descLines[1:] {
				b.WriteString(contIndent + line + "\n")
			}
		}
	}

	// 4. Notes / Supplementary format section
	if cfg.Notes != "" {
		b.WriteString("\n" + cfg.Notes + "\n")
	}

	// 5. Examples section
	if len(cfg.Examples) > 0 {
		b.WriteString("\nExamples:\n")
		for _, ex := range cfg.Examples {
			b.WriteString("  " + ex + "\n")
		}
	}

	return b.String()
}

// CmdHelper manages command documentation and formats standardized GNU help manuals.
type CmdHelper struct {
	Name        string
	Usage       []string
	Description string
	Notes       string
	Examples    []string
	ExtraFlags  []HelpOption // For synthetic or non-parsed flags (e.g. legacy options or manual help entries)
}

// SetUsage sets or replaces the usage synopsis lines.
func (h *CmdHelper) SetUsage(usage ...string) {
	h.Usage = usage
}

// SetDescription sets the command description.
func (h *CmdHelper) SetDescription(desc string) {
	h.Description = desc
}

// SetNotes sets notes / format descriptions.
func (h *CmdHelper) SetNotes(notes string) {
	h.Notes = notes
}

// AddExample adds an example usage line.
func (h *CmdHelper) AddExample(example string) {
	h.Examples = append(h.Examples, example)
}

// Format renders the complete help manual using the provided options slice.
func (h CmdHelper) Format(options []HelpOption) string {
	allOptions := make([]HelpOption, 0, len(options)+len(h.ExtraFlags))
	allOptions = append(allOptions, options...)
	allOptions = append(allOptions, h.ExtraFlags...)

	return FormatHelp(HelpConfig{
		Usage:       h.Usage,
		Description: h.Description,
		Options:     allOptions,
		Notes:       h.Notes,
		Examples:    h.Examples,
	})
}

