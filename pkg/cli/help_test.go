package cli

import (
	"strings"
	"testing"
)

func TestFormatHelp(t *testing.T) {
	cfg := HelpConfig{
		Usage: []string{
			"tool [OPTION]... [FILE]...",
			"tool OPTION",
		},
		Description: "A test utility for coreutils-go.",
		Options: []HelpOption{
			Opt('a', "all", "", "process all items"),
			Opt('n', "number", "NUM", "limit to NUM items"),
			Opt(0, "help", "", "display this help and exit"),
			Opt(0, "version", "", "output version information and exit"),
		},
		Notes: "Extra notes on syntax.",
		Examples: []string{
			"tool -a file.txt",
			"tool -n 10",
		},
	}

	help := FormatHelp(cfg)

	expectedSnippets := []string{
		"Usage: tool [OPTION]... [FILE]...\n  or:  tool OPTION\n",
		"Description:\nA test utility for coreutils-go.\n",
		"-a, --all",
		"-n, --number=NUM",
		"--help",
		"--version",
		"Extra notes on syntax.",
		"Examples:\n  tool -a file.txt\n  tool -n 10\n",
	}

	for _, snippet := range expectedSnippets {
		if !strings.Contains(help, snippet) {
			t.Errorf("FormatHelp output missing expected snippet %q:\n%s", snippet, help)
		}
	}
}

func TestCommand_IntegratedHelp(t *testing.T) {
	cmd := NewCommand("mytool", "mytool [OPTION]... [FILE]...", "mytool OPTION")
	cmd.SetDescription("A test utility for coreutils-go.")
	cmd.SetNotes("Extra notes on syntax.")
	cmd.AddExample("mytool -a file.txt")
	cmd.AddExample("mytool -n 10")

	cmd.FlagFunc('a', "all", "process all items", false, func(_ string) error { return nil })
	cmd.FlagArgFunc('n', "number", "NUM", "limit to NUM items", func(_ string) error { return nil })
	cmd.AddHelpOption(Opt(0, "help", "", "display this help and exit"))
	cmd.AddHelpOption(Opt(0, "version", "", "output version information and exit"))

	help := cmd.Help()

	for _, expected := range []string{
		"Usage: mytool [OPTION]... [FILE]...\n  or:  mytool OPTION\n",
		"Description:\nA test utility for coreutils-go.\n",
		"-a, --all",
		"-n, --number=NUM",
		"--help",
		"--version",
		"Extra notes on syntax.",
		"Examples:\n  mytool -a file.txt\n  mytool -n 10\n",
	} {
		if !strings.Contains(help, expected) {
			t.Errorf("cmd.Help() missing %q:\n%s", expected, help)
		}
	}
}

