package main

import (
	"bufio"
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
	tmpDir, err := os.MkdirTemp("", "yes-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	bin := filepath.Join(tmpDir, "yes.exe")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build yes binary: %v\n%s\n", err, out)
		os.Exit(1)
	}
	binaryPath = bin

	os.Exit(m.Run())
}

func runYesCLI(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func captureStdout(f func() error) (string, error) {
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

	execErr := f()

	w.Close()
	os.Stdout = oldStdout
	out := <-outChan
	r.Close()

	return out, execErr
}

func TestYes_DefaultStreaming(t *testing.T) {
	r, w := io.Pipe()

	errChan := make(chan error, 1)
	go func() {
		errChan <- runYes(w, nil)
	}()

	scanner := bufio.NewScanner(r)
	count := 0
	for count < 100 && scanner.Scan() {
		line := scanner.Text()
		if line != "y" {
			t.Fatalf("expected line 'y', got %q", line)
		}
		count++
	}

	_ = r.Close() // Simulate broken pipe / reader closing
	err := <-errChan
	if err == nil {
		t.Fatal("expected runYes to terminate with error upon reader close")
	}
}

func TestYes_CustomArgumentsStreaming(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "single argument",
			args:     []string{"hello"},
			expected: "hello",
		},
		{
			name:     "multiple arguments joined by space",
			args:     []string{"hello", "world", "123"},
			expected: "hello world 123",
		},
		{
			name:     "arguments with spaces and punctuation",
			args:     []string{"foo bar", "!@#$%^&*()"},
			expected: "foo bar !@#$%^&*()",
		},
		{
			name:     "unicode and emojis",
			args:     []string{"Olá mundo", "🚀🎉"},
			expected: "Olá mundo 🚀🎉",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, w := io.Pipe()

			errChan := make(chan error, 1)
			go func() {
				errChan <- runYes(w, tc.args)
			}()

			scanner := bufio.NewScanner(r)
			count := 0
			for count < 50 && scanner.Scan() {
				line := scanner.Text()
				if line != tc.expected {
					t.Fatalf("expected line %q, got %q", tc.expected, line)
				}
				count++
			}

			_ = r.Close()
			err := <-errChan
			if err == nil {
				t.Fatal("expected runYes to terminate on pipe closure")
			}
		})
	}
}

func TestYes_LargeArgument(t *testing.T) {
	// Pattern larger than the 64KB buffer
	largeArg := strings.Repeat("a", 70*1024)
	r, w := io.Pipe()

	errChan := make(chan error, 1)
	go func() {
		errChan <- runYes(w, []string{largeArg})
	}()

	scanner := bufio.NewScanner(r)
	buf := make([]byte, 100*1024)
	scanner.Buffer(buf, len(buf))

	if scanner.Scan() {
		line := scanner.Text()
		if line != largeArg {
			t.Fatalf("expected large line of length %d, got length %d", len(largeArg), len(line))
		}
	}

	_ = r.Close()
	_ = <-errChan
}

func TestYes_BuildBuffer(t *testing.T) {
	// Short line: should be repeated up to ~64KB
	line := []byte("y\n")
	buf := buildBuffer(line)
	if len(buf) == 0 || len(buf)%len(line) != 0 {
		t.Fatalf("expected buffer to be positive multiple of line length, got %d", len(buf))
	}
	if len(buf) < bufferSize-len(line) {
		t.Fatalf("expected buffer near %d bytes, got %d", bufferSize, len(buf))
	}

	// Line larger than bufferSize: should return unchanged
	largeLine := make([]byte, bufferSize+100)
	largeBuf := buildBuffer(largeLine)
	if len(largeBuf) != len(largeLine) {
		t.Fatalf("expected buffer to remain unchanged, got length %d", len(largeBuf))
	}
}

func TestYes_CLI_Help(t *testing.T) {
	for _, flag := range []string{"-h", "--help"} {
		t.Run(flag, func(t *testing.T) {
			stdout, stderr, err := runYesCLI(t, flag)
			if err != nil {
				t.Fatalf("expected exit 0 for %s, got error: %v, stderr: %q", flag, err, stderr)
			}
			if stderr != "" {
				t.Errorf("expected empty stderr, got %q", stderr)
			}
			for _, snippet := range []string{"Usage:", "yes", "Description:", "-h, --help", "-v, --version"} {
				if !strings.Contains(stdout, snippet) {
					t.Errorf("expected help output to contain %q, got:\n%s", snippet, stdout)
				}
			}
		})
	}
}

func TestYes_CLI_Version(t *testing.T) {
	for _, flag := range []string{"-v", "--version"} {
		t.Run(flag, func(t *testing.T) {
			stdout, stderr, err := runYesCLI(t, flag)
			if err != nil {
				t.Fatalf("expected exit 0 for %s, got error: %v, stderr: %q", flag, err, stderr)
			}
			if stderr != "" {
				t.Errorf("expected empty stderr, got %q", stderr)
			}
			if !strings.Contains(stdout, "yes (coreutils-go) 0.1.0") {
				t.Errorf("expected version output to contain 'yes (coreutils-go) 0.1.0', got %q", stdout)
			}
			if !strings.HasSuffix(stdout, "\n") {
				t.Errorf("expected version output to end with newline, got %q", stdout)
			}
		})
	}
}

func TestUnit_PrintHelp(t *testing.T) {
	help := printHelp()
	for _, snippet := range []string{"Usage:", "yes", "Description:", "-h, --help", "-v, --version"} {
		if !strings.Contains(help, snippet) {
			t.Errorf("printHelp() missing snippet %q", snippet)
		}
	}
}

func TestUnit_PrintVersion(t *testing.T) {
	out, err := captureStdout(func() error {
		return printVersion()
	})
	if err != nil {
		t.Fatalf("captureStdout error: %v", err)
	}
	if out != "yes (coreutils-go) 0.1.0\n" {
		t.Errorf("expected 'yes (coreutils-go) 0.1.0\\n', got %q", out)
	}
}

// limitWriter stops writing after totalBytes have been written and returns io.EOF.
type limitWriter struct {
	w       io.Writer
	remain  int64
}

func (lw *limitWriter) Write(p []byte) (n int, err error) {
	if lw.remain <= 0 {
		return 0, io.EOF
	}
	toWrite := int64(len(p))
	if toWrite > lw.remain {
		toWrite = lw.remain
	}
	n, err = lw.w.Write(p[:toWrite])
	lw.remain -= int64(n)
	if lw.remain <= 0 && err == nil {
		err = io.EOF
	}
	return n, err
}

func BenchmarkYes_Throughput(b *testing.B) {
	bytesPerIteration := int64(10 * 1024 * 1024) // 10 MB per iteration
	b.SetBytes(bytesPerIteration)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lw := &limitWriter{w: io.Discard, remain: bytesPerIteration}
		_ = runYes(lw, nil)
	}
}
