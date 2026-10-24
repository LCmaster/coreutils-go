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
