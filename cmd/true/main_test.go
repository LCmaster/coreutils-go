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
	tmpDir, err := os.MkdirTemp("", "true-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	bin := filepath.Join(tmpDir, "true.exe")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build true binary: %v\n%s\n", err, out)
		os.Exit(1)
	}
	binaryPath = bin

	os.Exit(m.Run())
}

func runTrue(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func runTrueWithStdin(t *testing.T, input string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

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

func TestTrue_BasicInvocation(t *testing.T) {
	stdout, stderr, err := runTrue(t)
	if err != nil {
		t.Fatalf("expected true to exit with 0, got error: %v, stderr: %q", err, stderr)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got %q", stderr)
	}
}

func TestTrue_PositionalArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "single argument",
			args: []string{"hello"},
		},
		{
			name: "multiple arguments",
			args: []string{"foo", "bar", "baz", "qux"},
		},
		{
			name: "empty string argument",
			args: []string{""},
		},
		{
			name: "multiple empty strings",
			args: []string{"", "", ""},
		},
		{
			name: "arguments with whitespace and tabs",
			args: []string{"hello world", "\tleading tab", "trailing space "},
		},
		{
			name: "numeric arguments",
			args: []string{"0", "1", "42", "99999"},
		},
		{
			name: "special characters and punctuation",
			args: []string{"!@#$%^&*()_+", "{}[];':\",.<>/?", "~`"},
		},
		{
			name: "unicode and emojis",
			args: []string{"Olá mundo", "こんにちは", "🚀🎉", "🌟✨"},
		},
		{
			name: "high count of arguments",
			args: func() []string {
				a := make([]string, 100)
				for i := range a {
					a[i] = fmt.Sprintf("arg-%d", i)
				}
				return a
			}(),
		},
		{
			name: "very long single argument",
			args: []string{strings.Repeat("long_text_", 100)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runTrue(t, tc.args...)
			if err != nil {
				t.Fatalf("expected true to exit with 0, got error: %v, stderr: %q", err, stderr)
			}
			if stdout != "" {
				t.Errorf("expected empty stdout, got %q", stdout)
			}
			if stderr != "" {
				t.Errorf("expected empty stderr, got %q", stderr)
			}
		})
	}
}

func TestTrue_IgnoredFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "unknown single short flag",
			args: []string{"-a"},
		},
		{
			name: "unknown grouped short flags",
			args: []string{"-xyz"},
		},
		{
			name: "unknown long flag",
			args: []string{"--custom-flag"},
		},
		{
			name: "unknown long flag with value",
			args: []string{"--custom=value"},
		},
		{
			name: "bare single dash",
			args: []string{"-"},
		},
		{
			name: "bare double dash",
			args: []string{"--"},
		},
		{
			name: "double dash followed by arguments",
			args: []string{"--", "foo", "bar"},
		},
		{
			name: "common unix flags ignored",
			args: []string{"-f", "-r", "-l", "-e", "-E", "-n", "-s"},
		},
		{
			name: "flags mixed with positional arguments",
			args: []string{"-x", "foo", "--verbose", "bar", "-z"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runTrue(t, tc.args...)
			if err != nil {
				t.Fatalf("expected true to exit with 0, got error: %v, stderr: %q", err, stderr)
			}
			if stdout != "" {
				t.Errorf("expected empty stdout, got %q", stdout)
			}
			if stderr != "" {
				t.Errorf("expected empty stderr, got %q", stderr)
			}
		})
	}
}

func TestTrue_VersionFlag(t *testing.T) {
	for _, flag := range []string{"-v", "--version"} {
		t.Run(flag, func(t *testing.T) {
			stdout, stderr, err := runTrue(t, flag)
			if err != nil {
				t.Fatalf("expected exit 0, got error: %v, stderr: %q", err, stderr)
			}
			if stderr != "" {
				t.Errorf("expected empty stderr, got %q", stderr)
			}
			if !strings.Contains(stdout, "true (coreutils-go) 0.1.0") {
				t.Errorf("expected version output to contain 'true (coreutils-go) 0.1.0', got %q", stdout)
			}
			if !strings.HasSuffix(stdout, "\n") {
				t.Errorf("expected version output to end with newline, got %q", stdout)
			}
		})
	}
}

func TestTrue_HelpFlag(t *testing.T) {
	expectedSnippets := []string{
		"Usage:",
		"true [ignored command line arguments]",
		"Description:",
		"Exit with a status code indicating success.",
		"-h, --help",
		"-v, --version",
	}

	for _, flag := range []string{"-h", "--help"} {
		t.Run(flag, func(t *testing.T) {
			stdout, stderr, err := runTrue(t, flag)
			if err != nil {
				t.Fatalf("expected exit 0, got error: %v, stderr: %q", err, stderr)
			}
			if stderr != "" {
				t.Errorf("expected empty stderr, got %q", stderr)
			}
			for _, snippet := range expectedSnippets {
				if !strings.Contains(stdout, snippet) {
					t.Errorf("expected help output to contain %q, got:\n%s", snippet, stdout)
				}
			}
		})
	}
}

func TestTrue_SoleArgumentRule(t *testing.T) {
	// GNU specification: --help and --version are ONLY recognized when they are the sole argument.
	// When accompanied by other arguments, they are ignored and true exits 0 silently.
	tests := []struct {
		name string
		args []string
	}{
		{name: "version followed by arg", args: []string{"--version", "extra_arg"}},
		{name: "version preceded by arg", args: []string{"extra_arg", "--version"}},
		{name: "short version followed by arg", args: []string{"-v", "extra_arg"}},
		{name: "help followed by arg", args: []string{"--help", "extra_arg"}},
		{name: "help preceded by arg", args: []string{"extra_arg", "--help"}},
		{name: "short help followed by arg", args: []string{"-h", "extra_arg"}},
		{name: "both help and version together", args: []string{"--help", "--version"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runTrue(t, tc.args...)
			if err != nil {
				t.Fatalf("expected exit 0, got error: %v, stderr: %q", err, stderr)
			}
			if stdout != "" {
				t.Errorf("expected empty stdout under sole argument rule, got %q", stdout)
			}
			if stderr != "" {
				t.Errorf("expected empty stderr under sole argument rule, got %q", stderr)
			}
		})
	}
}

func TestTrue_StdinIgnored(t *testing.T) {
	inputs := []string{
		"hello world\n",
		strings.Repeat("data line\n", 100),
		"\x00\x01\x02\xFF",
	}

	for i, input := range inputs {
		t.Run(fmt.Sprintf("stdin_case_%d", i), func(t *testing.T) {
			stdout, stderr, err := runTrueWithStdin(t, input)
			if err != nil {
				t.Fatalf("expected exit 0 with stdin input, got error: %v, stderr: %q", err, stderr)
			}
			if stdout != "" {
				t.Errorf("expected empty stdout, got %q", stdout)
			}
			if stderr != "" {
				t.Errorf("expected empty stderr, got %q", stderr)
			}
		})
	}
}

func TestUnit_PrintHelp(t *testing.T) {
	help := newCommand().Help()
	expectedSnippets := []string{
		"Usage:",
		"true [ignored command line arguments]",
		"Description:",
		"Exit with a status code indicating success.",
		"-h, --help",
		"-v, --version",
	}

	for _, snippet := range expectedSnippets {
		if !strings.Contains(help, snippet) {
			t.Errorf("newCommand().Help() missing snippet %q", snippet)
		}
	}
}

func TestUnit_PrintVersion(t *testing.T) {
	ver := newCommand().Version()
	if ver != "true (coreutils-go) 0.1.0" {
		t.Errorf("expected 'true (coreutils-go) 0.1.0', got %q", ver)
	}
}


