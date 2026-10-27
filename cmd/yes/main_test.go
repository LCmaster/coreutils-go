package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

	name := "yes"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	bin := filepath.Join(tmpDir, name)
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build yes binary: %v\n%s\n", err, out)
		_ = os.RemoveAll(tmpDir)
		os.Exit(1)
	}
	binaryPath = bin

	code := m.Run()
	_ = os.RemoveAll(tmpDir)
	os.Exit(code)
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

// readNLines reads up to n lines from r using bufio.Scanner and verifies each line equals expected.
func readNLines(t *testing.T, r io.Reader, n int, expected string) {
	t.Helper()
	scanner := bufio.NewScanner(r)
	count := 0
	for count < n && scanner.Scan() {
		line := scanner.Text()
		if line != expected {
			t.Fatalf("line %d: expected %q, got %q", count, expected, line)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error after %d lines: %v", count, err)
	}
	if count != n {
		t.Fatalf("expected to read %d lines, but only read %d", n, count)
	}
}

func TestYes_DefaultStreaming(t *testing.T) {
	r, w := io.Pipe()

	errChan := make(chan error, 1)
	go func() {
		errChan <- runYes(w, nil)
	}()

	readNLines(t, r, 200, "y")

	_ = r.Close() // Simulate consumer closing stream
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
			name:     "empty string argument produces empty line",
			args:     []string{""},
			expected: "",
		},
		{
			name:     "multiple empty string arguments separated by space",
			args:     []string{"", ""},
			expected: " ",
		},
		{
			name:     "arguments with spaces and punctuation",
			args:     []string{"foo bar", "!@#$%^&*()"},
			expected: "foo bar !@#$%^&*()",
		},
		{
			name:     "arguments with quotes and tabs",
			args:     []string{`"quoted"`, "'single'", "tab\tspace"},
			expected: "\"quoted\" 'single' tab\tspace",
		},
		{
			name:     "numeric arguments",
			args:     []string{"0", "1", "-42", "3.14159"},
			expected: "0 1 -42 3.14159",
		},
		{
			name:     "unicode and emojis",
			args:     []string{"Olá mundo", "こんにちは", "🚀🎉🔥"},
			expected: "Olá mundo こんにちは 🚀🎉🔥",
		},
		{
			name: "many arguments joined",
			args: func() []string {
				a := make([]string, 50)
				for i := range a {
					a[i] = fmt.Sprintf("val%d", i)
				}
				return a
			}(),
			expected: func() string {
				a := make([]string, 50)
				for i := range a {
					a[i] = fmt.Sprintf("val%d", i)
				}
				return strings.Join(a, " ")
			}(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, w := io.Pipe()

			errChan := make(chan error, 1)
			go func() {
				errChan <- runYes(w, tc.args)
			}()

			readNLines(t, r, 50, tc.expected)

			_ = r.Close()
			err := <-errChan
			if err == nil {
				t.Fatal("expected runYes to terminate on pipe closure")
			}
		})
	}
}

func TestYes_NonDivisibleLineLengths(t *testing.T) {
	// Test various pattern lengths that do not evenly divide bufferSize (64KB)
	// to ensure no partial or corrupted lines are emitted across buffer boundaries.
	for _, length := range []int{3, 7, 11, 13, 17, 31, 101, 257, 1023} {
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			pattern := strings.Repeat("x", length-1) // with newline, length is exact
			r, w := io.Pipe()

			errChan := make(chan error, 1)
			go func() {
				errChan <- runYes(w, []string{pattern})
			}()

			readNLines(t, r, 200, pattern)

			_ = r.Close()
			_ = <-errChan
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

	for i := 0; i < 5; i++ {
		if !scanner.Scan() {
			t.Fatalf("scanner failed at iteration %d: %v", i, scanner.Err())
		}
		line := scanner.Text()
		if line != largeArg {
			t.Fatalf("expected large line of length %d, got length %d", len(largeArg), len(line))
		}
	}

	_ = r.Close()
	_ = <-errChan
}

func TestYes_BuildBufferIntegrity(t *testing.T) {
	// 1. Short line: must be repeated as full lines only
	line := []byte("hello\n")
	buf := buildBuffer(line)
	if len(buf) == 0 {
		t.Fatal("expected non-empty buffer")
	}
	if len(buf)%len(line) != 0 {
		t.Fatalf("buffer size %d is not a clean multiple of line length %d", len(buf), len(line))
	}
	// Verify every segment equals the line
	for i := 0; i < len(buf); i += len(line) {
		segment := buf[i : i+len(line)]
		if !bytes.Equal(segment, line) {
			t.Fatalf("corrupted segment at offset %d: %q", i, segment)
		}
	}

	// 2. Exact bufferSize line
	exactLine := append(bytes.Repeat([]byte("a"), bufferSize-1), '\n')
	exactBuf := buildBuffer(exactLine)
	if len(exactBuf) != len(exactLine) {
		t.Fatalf("expected exact buffer of length %d, got %d", len(exactLine), len(exactBuf))
	}

	// 3. Oversized line
	largeLine := append(bytes.Repeat([]byte("b"), bufferSize+1000), '\n')
	largeBuf := buildBuffer(largeLine)
	if len(largeBuf) != len(largeLine) {
		t.Fatalf("expected oversized buffer of length %d, got %d", len(largeLine), len(largeBuf))
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

func TestYes_CLI_SubprocessPipe(t *testing.T) {
	// Run actual compiled binary as a child process and read from stdout pipe
	cmd := exec.Command(binaryPath, "stream_test")
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to open stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start yes process: %v", err)
	}

	readNLines(t, stdoutPipe, 100, "stream_test")

	// Close stdout pipe to simulate broken pipe and verify process exits
	_ = stdoutPipe.Close()
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
}

func TestYes_CLI_MultipleArgsWithFlags(t *testing.T) {
	// When --help or --version is accompanied by other arguments,
	// it should be treated as text to output rather than triggering help/version.
	cmd := exec.Command(binaryPath, "--help", "extra_arg")
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to open stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start yes process: %v", err)
	}

	readNLines(t, stdoutPipe, 20, "--help extra_arg")

	_ = stdoutPipe.Close()
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
}

func TestUnit_PrintHelp(t *testing.T) {
	help := newCommand().Help()
	for _, snippet := range []string{"Usage:", "yes", "Description:", "-h, --help", "-v, --version"} {
		if !strings.Contains(help, snippet) {
			t.Errorf("newCommand().Help() missing snippet %q", snippet)
		}
	}
}

func TestUnit_PrintVersion(t *testing.T) {
	ver := newCommand().Version()
	if ver != "yes (coreutils-go) 0.1.0" {
		t.Errorf("expected 'yes (coreutils-go) 0.1.0', got %q", ver)
	}
}


// limitWriter stops writing after totalBytes have been written and returns io.EOF.
type limitWriter struct {
	w      io.Writer
	remain int64
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

func BenchmarkYes_Throughput_Default(b *testing.B) {
	bytesPerIteration := int64(10 * 1024 * 1024) // 10 MB per iteration
	b.SetBytes(bytesPerIteration)

	for b.Loop() {
		lw := &limitWriter{w: io.Discard, remain: bytesPerIteration}
		_ = runYes(lw, nil)
	}
}

func BenchmarkYes_Throughput_ShortWord(b *testing.B) {
	bytesPerIteration := int64(10 * 1024 * 1024)
	b.SetBytes(bytesPerIteration)

	for b.Loop() {
		lw := &limitWriter{w: io.Discard, remain: bytesPerIteration}
		_ = runYes(lw, []string{"confirm"})
	}
}

func BenchmarkYes_Throughput_LongSentence(b *testing.B) {
	bytesPerIteration := int64(10 * 1024 * 1024)
	b.SetBytes(bytesPerIteration)

	sentence := []string{"The", "quick", "brown", "fox", "jumps", "over", "the", "lazy", "dog"}
	for b.Loop() {
		lw := &limitWriter{w: io.Discard, remain: bytesPerIteration}
		_ = runYes(lw, sentence)
	}
}
