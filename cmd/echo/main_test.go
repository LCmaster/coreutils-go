package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	// Build binary to a temporary directory for CLI end-to-end integration tests
	tmpDir, err := os.MkdirTemp("", "echo-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	bin := filepath.Join(tmpDir, "echo.exe")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build echo binary: %v\n%s\n", err, out)
		os.Exit(1)
	}
	binaryPath = bin

	os.Exit(m.Run())
}

// runEcho runs the compiled echo binary with the given arguments and returns stdout, stderr, and exit error.
func runEcho(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// captureStdout redirects os.Stdout during f() and returns the captured output.
func captureStdout(f func()) (string, error) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w

	outChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outChan <- buf.String()
	}()

	f()

	w.Close()
	os.Stdout = oldStdout
	out := <-outChan
	r.Close()

	return out, nil
}

func TestEcho_BasicOutput(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "no arguments outputs a newline",
			args:     []string{},
			expected: "\n",
		},
		{
			name:     "single argument",
			args:     []string{"hello"},
			expected: "hello\n",
		},
		{
			name:     "two arguments separated by space",
			args:     []string{"hello", "world"},
			expected: "hello world\n",
		},
		{
			name:     "multiple arguments",
			args:     []string{"foo", "bar", "baz", "qux"},
			expected: "foo bar baz qux\n",
		},
		{
			name:     "empty string argument",
			args:     []string{""},
			expected: "\n",
		},
		{
			name:     "multiple empty string arguments",
			args:     []string{"", ""},
			expected: " \n",
		},
		{
			name:     "arguments with internal spaces",
			args:     []string{"hello world", "foo bar"},
			expected: "hello world foo bar\n",
		},
		{
			name:     "arguments with special characters and symbols",
			args:     []string{"123", "!@#$%^&*()_+", "{}[];':\",.<>/?"},
			expected: "123 !@#$%^&*()_+ {}[];':\",.<>/?\n",
		},
		{
			name:     "unicode and emojis",
			args:     []string{"Olá mundo", "こんにちは", "🚀🎉"},
			expected: "Olá mundo こんにちは 🚀🎉\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runEcho(t, tc.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v, stderr: %q", err, stderr)
			}
			if stdout != tc.expected {
				t.Errorf("expected stdout %q, got %q", tc.expected, stdout)
			}
		})
	}
}

func TestEcho_NewlineFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "-n alone outputs empty string",
			args:     []string{"-n"},
			expected: "",
		},
		{
			name:     "-n with single argument",
			args:     []string{"-n", "hello"},
			expected: "hello",
		},
		{
			name:     "-n with multiple arguments",
			args:     []string{"-n", "hello", "world"},
			expected: "hello world",
		},
		{
			name:     "-n with empty string",
			args:     []string{"-n", ""},
			expected: "",
		},
		{
			name:     "-n with space inside argument",
			args:     []string{"-n", "Prompt: "},
			expected: "Prompt: ",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runEcho(t, tc.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v, stderr: %q", err, stderr)
			}
			if stdout != tc.expected {
				t.Errorf("expected stdout %q, got %q", tc.expected, stdout)
			}
		})
	}
}

func TestEcho_DisableEscapeFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "default without -e does not interpret newline escape",
			args:     []string{`hello\nworld`},
			expected: "hello\\nworld\n",
		},
		{
			name:     "default without -e does not interpret tab escape",
			args:     []string{`tab\ttest`},
			expected: "tab\\ttest\n",
		},
		{
			name:     "-E explicitly disables backslash interpretation",
			args:     []string{"-E", `hello\nworld`},
			expected: "hello\\nworld\n",
		},
		{
			name:     "-E with literal backslash",
			args:     []string{"-E", `\\`},
			expected: "\\\\\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runEcho(t, tc.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v, stderr: %q", err, stderr)
			}
			if stdout != tc.expected {
				t.Errorf("expected stdout %q, got %q", tc.expected, stdout)
			}
		})
	}
}

func TestEcho_EnableEscapeFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "newline escape sequence",
			args:     []string{"-e", `hello\nworld`},
			expected: "hello\nworld\n",
		},
		{
			name:     "horizontal tab escape sequence",
			args:     []string{"-e", `col1\tcol2`},
			expected: "col1\tcol2\n",
		},
		{
			name:     "backslash escape sequence",
			args:     []string{"-e", `path\\to\\file`},
			expected: "path\\to\\file\n",
		},
		{
			name:     "alert (BEL) escape sequence",
			args:     []string{"-e", `alert\a`},
			expected: "alert\a\n",
		},
		{
			name:     "backspace escape sequence",
			args:     []string{"-e", `back\bspace`},
			expected: "back\bspace\n",
		},
		{
			name:     "form feed escape sequence",
			args:     []string{"-e", `form\ffeed`},
			expected: "form\ffeed\n",
		},
		{
			name:     "carriage return escape sequence",
			args:     []string{"-e", `line\r`},
			expected: "line\r\n",
		},
		{
			name:     "vertical tab escape sequence",
			args:     []string{"-e", `vert\vtab`},
			expected: "vert\vtab\n",
		},
		{
			name:     "escape character \\e",
			args:     []string{"-e", `\e[32mOK\e[0m`},
			expected: "\x1b[32mOK\x1b[0m\n",
		},
		{
			name:     "octal escape sequence",
			args:     []string{"-e", `\0101`},
			expected: "A\n",
		},
		{
			name:     "hexadecimal escape sequence",
			args:     []string{"-e", `\x41`},
			expected: "A\n",
		},
		{
			name:     "multiple escape sequences in one string",
			args:     []string{"-e", `a\tb\nc\td`},
			expected: "a\tb\nc\td\n",
		},
		{
			name:     "unrecognized escape kept literally",
			args:     []string{"-e", `\z\y`},
			expected: "\\z\\y\n",
		},
		{
			name:     "trailing backslash preserved",
			args:     []string{"-e", `trailing\`},
			expected: "trailing\\\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runEcho(t, tc.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v, stderr: %q", err, stderr)
			}
			if stdout != tc.expected {
				t.Errorf("expected stdout %q, got %q", tc.expected, stdout)
			}
		})
	}
}

func TestEcho_CancelOutputEscape(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "\\c halts immediately and suppresses trailing newline",
			args:     []string{"-e", `hello\cworld`},
			expected: "hello",
		},
		{
			name:     "\\c at start halts immediately",
			args:     []string{"-e", `\chello`},
			expected: "",
		},
		{
			name:     "\\c halts across subsequent arguments",
			args:     []string{"-e", `first\c`, "second", "third"},
			expected: "first",
		},
		{
			name:     "\\c in second argument",
			args:     []string{"-e", "first", `second\chalt`, "third"},
			expected: "first second",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runEcho(t, tc.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v, stderr: %q", err, stderr)
			}
			if stdout != tc.expected {
				t.Errorf("expected stdout %q, got %q", tc.expected, stdout)
			}
		})
	}
}

func TestEcho_QuotesWithEscape(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "double quotes preserved with -e newline",
			args:     []string{"-e", `hello "world"\ntest`},
			expected: "hello \"world\"\ntest\n",
		},
		{
			name:     "single quotes preserved with -e tab",
			args:     []string{"-e", `it's\ta test`},
			expected: "it's\ta test\n",
		},
		{
			name:     "mixed quotes and multiple escapes",
			args:     []string{"-e", `"quoted"\t'single'\nend`},
			expected: "\"quoted\"\t'single'\nend\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runEcho(t, tc.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v, stderr: %q", err, stderr)
			}
			if stdout != tc.expected {
				t.Errorf("expected stdout %q, got %q", tc.expected, stdout)
			}
		})
	}
}

func TestEcho_FlagPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "-e then -E: -E wins (escapes disabled)",
			args:     []string{"-e", "-E", `hello\nworld`},
			expected: "hello\\nworld\n",
		},
		{
			name:     "-E then -e: -e wins (escapes enabled)",
			args:     []string{"-E", "-e", `hello\nworld`},
			expected: "hello\nworld\n",
		},
		{
			name:     "grouped -eE: E is last, escapes disabled",
			args:     []string{"-eE", `hello\nworld`},
			expected: "hello\\nworld\n",
		},
		{
			name:     "grouped -Ee: e is last, escapes enabled",
			args:     []string{"-Ee", `hello\nworld`},
			expected: "hello\nworld\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runEcho(t, tc.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v, stderr: %q", err, stderr)
			}
			if stdout != tc.expected {
				t.Errorf("expected stdout %q, got %q", tc.expected, stdout)
			}
		})
	}
}

func TestEcho_CombinedFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "-n and -e as separate flags",
			args:     []string{"-n", "-e", `hello\nworld`},
			expected: "hello\nworld",
		},
		{
			name:     "-e and -n as separate flags",
			args:     []string{"-e", "-n", `hello\nworld`},
			expected: "hello\nworld",
		},
		{
			name:     "-ne as combined short flag",
			args:     []string{"-ne", `hello\nworld`},
			expected: "hello\nworld",
		},
		{
			name:     "-en as combined short flag",
			args:     []string{"-en", `hello\nworld`},
			expected: "hello\nworld",
		},
		{
			name:     "-n -E preserves escapes without trailing newline",
			args:     []string{"-n", "-E", `hello\nworld`},
			expected: "hello\\nworld",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runEcho(t, tc.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v, stderr: %q", err, stderr)
			}
			if stdout != tc.expected {
				t.Errorf("expected stdout %q, got %q", tc.expected, stdout)
			}
		})
	}
}

func TestEcho_VersionFlag(t *testing.T) {
	for _, flag := range []string{"-v", "--version"} {
		t.Run(flag, func(t *testing.T) {
			stdout, stderr, err := runEcho(t, flag)
			if err != nil {
				t.Fatalf("unexpected error: %v, stderr: %q", err, stderr)
			}
			if !strings.Contains(stdout, "version 0.1.0") {
				t.Errorf("expected version output to contain 'version 0.1.0', got %q", stdout)
			}
			if !strings.HasSuffix(stdout, "\n") {
				t.Errorf("expected version output to end with newline, got %q", stdout)
			}
		})
	}
}

func TestEcho_HelpFlag(t *testing.T) {
	for _, flag := range []string{"-h", "--help"} {
		t.Run(flag, func(t *testing.T) {
			stdout, stderr, err := runEcho(t, flag)
			if err != nil {
				t.Fatalf("unexpected error: %v, stderr: %q", err, stderr)
			}
			for _, expectedSnippet := range []string{
				"Usage:",
				"echo [OPTIONS]... [STRING]...",
				"Description:",
				"-n",
				"-e",
				"-E",
				"-h, --help",
				"-v, --version",
			} {
				if !strings.Contains(stdout, expectedSnippet) {
					t.Errorf("help output missing snippet %q, got:\n%s", expectedSnippet, stdout)
				}
			}
		})
	}
}



func TestUnit_PrintHelp(t *testing.T) {
	help := printHelp()
	expectedSubstrings := []string{
		"Usage:",
		"echo [OPTIONS]... [STRING]...",
		"Description:",
		"\\n\t\tnew line",
		"\\t\t\thorizontal tab",
		"\\0NNN\t\tbyte with octal value NNN",
		"\\xHH\t\tbyte with hexadecimal value HH",
	}

	for _, sub := range expectedSubstrings {
		if !strings.Contains(help, sub) {
			t.Errorf("printHelp() missing expected substring %q", sub)
		}
	}
}

func TestUnit_PrintVersion(t *testing.T) {
	out, err := captureStdout(func() {
		printVersion()
	})
	if err != nil {
		t.Fatalf("captureStdout error: %v", err)
	}
	if out != "version 0.1.0\n" {
		t.Errorf("expected 'version 0.1.0\\n', got %q", out)
	}
}
