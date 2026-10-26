package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "head-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	name := "head-test"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	bin := filepath.Join(tmpDir, name)
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build head binary: %v\n%s\n", err, out)
		_ = os.RemoveAll(tmpDir)
		os.Exit(1)
	}
	binaryPath = bin

	code := m.Run()
	_ = os.RemoveAll(tmpDir)
	os.Exit(code)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func runHead(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	return runHeadStdin(t, "", args...)
}

func runHeadStdin(t *testing.T, stdin string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run head: %v", err)
		}
	}
	return stdout.String(), stderr.String(), code
}

func writeFile(t *testing.T, name string, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
	return path
}

func executeMemory(args []string, stdin io.Reader) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd := newCommand()
	err := cmd.Execute(args, stdin, &stdout, &stderr)
	return stdout.String(), stderr.String(), err
}

type brokenPipeWriter struct{}

func (b *brokenPipeWriter) Write(_ []byte) (int, error) {
	return 0, syscall.EPIPE
}

// ---------------------------------------------------------------------------
// 1. Unit Tests: Metadata, Flags, & parseNum
// ---------------------------------------------------------------------------

func TestUnit_Help(t *testing.T) {
	help := newCommand().Help()
	snippets := []string{
		"Usage:",
		"head [OPTION]... [FILE]...",
		"Description:",
		"Print the first 10 lines of each FILE to standard output.",
		"-c, --bytes",
		"-n, --lines",
		"-q, --quiet",
		"-v, --verbose",
		"-z, --zero-terminated",
		"NUM may have a multiplier suffix",
	}
	for _, snip := range snippets {
		if !strings.Contains(help, snip) {
			t.Errorf("help missing snippet %q\nFull help:\n%s", snip, help)
		}
	}
}

func TestUnit_ParseNum(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		wantErr  bool
	}{
		{"1", 1, false},
		{"10", 10, false},
		{"100", 100, false},
		{"+5", 5, false},
		{"+100", 100, false},
		{"-1", -1, false},
		{"-10", -10, false},
		{"0", 0, false},
		{"+0", 0, false},
		{"-0", 0, false},

		// Suffixes
		{"1b", 512, false},
		{"2b", 1024, false},
		{"1k", 1024, false},
		{"1K", 1024, false},
		{"1KiB", 1024, false},
		{"1kB", 1000, false},
		{"1KB", 1000, false},
		{"1m", 1024 * 1024, false},
		{"1M", 1024 * 1024, false},
		{"1MiB", 1024 * 1024, false},
		{"1MB", 1000 * 1000, false},
		{"1g", 1024 * 1024 * 1024, false},
		{"1G", 1024 * 1024 * 1024, false},
		{"1GiB", 1024 * 1024 * 1024, false},
		{"1GB", 1000 * 1000 * 1000, false},
		{"-2K", -2048, false},
		{"+2K", 2048, false},

		// Errors
		{"", 0, true},
		{"+", 0, true},
		{"-", 0, true},
		{"abc", 0, true},
		{"10abc", 0, true},
		{"10.5", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseNum(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseNum(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("parseNum(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestUnit_InvalidNumberFlags(t *testing.T) {
	_, stderr, code := runHeadStdin(t, "data", "-n", "invalid")
	if code == 0 {
		t.Errorf("expected non-zero exit code for -n invalid, got %d", code)
	}
	if !strings.Contains(stderr, "invalid number of lines: 'invalid'") {
		t.Errorf("unexpected stderr: %s", stderr)
	}

	_, stderr, code = runHeadStdin(t, "data", "-c", "invalid")
	if code == 0 {
		t.Errorf("expected non-zero exit code for -c invalid, got %d", code)
	}
	if !strings.Contains(stderr, "invalid number of bytes: 'invalid'") {
		t.Errorf("unexpected stderr: %s", stderr)
	}
}

func TestUnit_ExecuteMemory_Direct(t *testing.T) {
	// Tests in-process execution paths for coverage and fast execution
	t.Run("default lines", func(t *testing.T) {
		out, errOut, err := executeMemory([]string{}, strings.NewReader("1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n"))
		if err != nil || errOut != "" || out != "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n" {
			t.Errorf("unexpected output: out=%q, errOut=%q, err=%v", out, errOut, err)
		}
	})

	t.Run("positive lines", func(t *testing.T) {
		out, _, err := executeMemory([]string{"-n", "2"}, strings.NewReader("a\nb\nc\n"))
		if err != nil || out != "a\nb\n" {
			t.Errorf("unexpected out: %q, err: %v", out, err)
		}
	})

	t.Run("negative lines", func(t *testing.T) {
		out, _, err := executeMemory([]string{"-n", "-1"}, strings.NewReader("a\nb\nc\n"))
		if err != nil || out != "a\nb\n" {
			t.Errorf("unexpected out: %q, err: %v", out, err)
		}
	})

	t.Run("positive bytes", func(t *testing.T) {
		out, _, err := executeMemory([]string{"-c", "3"}, strings.NewReader("abcdef"))
		if err != nil || out != "abc" {
			t.Errorf("unexpected out: %q, err: %v", out, err)
		}
	})

	t.Run("negative bytes", func(t *testing.T) {
		out, _, err := executeMemory([]string{"-c", "-2"}, strings.NewReader("abcdef"))
		if err != nil || out != "abcd" {
			t.Errorf("unexpected out: %q, err: %v", out, err)
		}
	})

	t.Run("zero terminated", func(t *testing.T) {
		out, _, err := executeMemory([]string{"-z", "-n", "1"}, strings.NewReader("foo\x00bar\x00"))
		if err != nil || out != "foo\x00" {
			t.Errorf("unexpected out: %q, err: %v", out, err)
		}
	})

	t.Run("file with headers", func(t *testing.T) {
		f1 := writeFile(t, "m1.txt", "line1\n")
		f2 := writeFile(t, "m2.txt", "line2\n")
		out, _, err := executeMemory([]string{f1, f2}, nil)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		expected := fmt.Sprintf("==> %s <==\nline1\n\n==> %s <==\nline2\n", f1, f2)
		if out != expected {
			t.Errorf("out = %q, want %q", out, expected)
		}
	})

	t.Run("quiet mode multiple files", func(t *testing.T) {
		f1 := writeFile(t, "mq1.txt", "a\n")
		f2 := writeFile(t, "mq2.txt", "b\n")
		out, _, err := executeMemory([]string{"-q", f1, f2}, nil)
		if err != nil || out != "a\nb\n" {
			t.Errorf("out = %q, err = %v", out, err)
		}
	})

	t.Run("verbose mode single file", func(t *testing.T) {
		f1 := writeFile(t, "mv1.txt", "hi\n")
		out, _, err := executeMemory([]string{"-v", f1}, nil)
		if err != nil || out != fmt.Sprintf("==> %s <==\nhi\n", f1) {
			t.Errorf("out = %q, err = %v", out, err)
		}
	})

	t.Run("missing file returns error", func(t *testing.T) {
		_, errOut, err := executeMemory([]string{"nonexistent_direct.txt"}, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(errOut, "cannot open 'nonexistent_direct.txt'") {
			t.Errorf("expected diagnostic in stderr, got: %s", errOut)
		}
	})

	t.Run("non-seekable negative bytes", func(t *testing.T) {
		pr, pw := io.Pipe()
		go func() {
			_, _ = pw.Write([]byte("abcdefghij"))
			_ = pw.Close()
		}()
		var out bytes.Buffer
		err := copyAllExceptLastBytes(&out, pr, 3)
		if err != nil || out.String() != "abcdefg" {
			t.Errorf("out = %q, err = %v", out.String(), err)
		}
	})

	t.Run("zero or negative k bounds", func(t *testing.T) {
		var out bytes.Buffer
		if err := copyFirstLines(&out, strings.NewReader("hi\n"), '\n', 0); err != nil {
			t.Fatal(err)
		}
		if err := copyAllExceptLastLines(&out, strings.NewReader("hi\n"), '\n', 0); err != nil {
			t.Fatal(err)
		}
		if err := copyFirstBytes(&out, strings.NewReader("hi\n"), 0); err != nil {
			t.Fatal(err)
		}
		if err := copyAllExceptLastBytes(&out, strings.NewReader("hi\n"), 0); err != nil {
			t.Fatal(err)
		}
	})
}

// ---------------------------------------------------------------------------
// 2. Default 10 Lines
// ---------------------------------------------------------------------------

func TestHead_DefaultTenLines_Stdin(t *testing.T) {
	var sb strings.Builder
	for i := 1; i <= 20; i++ {
		fmt.Fprintf(&sb, "line %02d\n", i)
	}

	stdout, stderr, code := runHeadStdin(t, sb.String())
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}
	expected := "line 01\nline 02\nline 03\nline 04\nline 05\nline 06\nline 07\nline 08\nline 09\nline 10\n"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestHead_DefaultTenLines_FewerThanTen(t *testing.T) {
	input := "row 1\nrow 2\nrow 3\n"
	stdout, stderr, code := runHeadStdin(t, input)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}
	if stdout != input {
		t.Errorf("stdout = %q, want %q", stdout, input)
	}
}

func TestHead_DefaultTenLines_ExactlyTen(t *testing.T) {
	input := "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n"
	stdout, stderr, code := runHeadStdin(t, input)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}
	if stdout != input {
		t.Errorf("stdout = %q, want %q", stdout, input)
	}
}

func TestHead_EmptyInput(t *testing.T) {
	stdout, stderr, code := runHeadStdin(t, "")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
}

func TestHead_SingleLineNoTrailingNewline(t *testing.T) {
	input := "hello world without newline"
	stdout, stderr, code := runHeadStdin(t, input)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}
	if stdout != input {
		t.Errorf("stdout = %q, want %q", stdout, input)
	}
}

func TestHead_DefaultTenLines_File(t *testing.T) {
	var sb strings.Builder
	for i := 1; i <= 15; i++ {
		fmt.Fprintf(&sb, "entry %d\n", i)
	}
	f := writeFile(t, "sample.txt", sb.String())

	stdout, stderr, code := runHead(t, f)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}
	expected := "entry 1\nentry 2\nentry 3\nentry 4\nentry 5\nentry 6\nentry 7\nentry 8\nentry 9\nentry 10\n"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

// ---------------------------------------------------------------------------
// 3. Custom Line Counts (-n NUM and legacy -NUM)
// ---------------------------------------------------------------------------

func TestHead_Lines_Zero(t *testing.T) {
	stdout, _, code := runHeadStdin(t, "line1\nline2\n", "-n", "0")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
}

func TestHead_Lines_Positive(t *testing.T) {
	input := "1\n2\n3\n4\n5\n"
	stdout, _, code := runHeadStdin(t, input, "-n", "3")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	expected := "1\n2\n3\n"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestHead_Lines_ExplicitPlus(t *testing.T) {
	input := "1\n2\n3\n4\n5\n"
	stdout, _, code := runHeadStdin(t, input, "-n", "+2")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	expected := "1\n2\n"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestHead_Lines_NumericShorthand(t *testing.T) {
	input := "alpha\nbeta\ngamma\ndelta\n"
	stdout, _, code := runHeadStdin(t, input, "-2")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	expected := "alpha\nbeta\n"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}

	stdout, _, code = runHeadStdin(t, input, "-1")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout != "alpha\n" {
		t.Errorf("stdout = %q, want %q", stdout, "alpha\n")
	}
}

func TestHead_Lines_Suffix(t *testing.T) {
	var sb strings.Builder
	for i := 1; i <= 2000; i++ {
		fmt.Fprintf(&sb, "item %d\n", i)
	}
	stdout, _, code := runHeadStdin(t, sb.String(), "-n", "1k")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
	if len(lines) != 1024 {
		t.Errorf("expected 1024 lines, got %d", len(lines))
	}
}

func TestHead_Lines_FlagOverride(t *testing.T) {
	input := "1\n2\n3\n4\n5\n"
	// Later flag should take precedence
	stdout, _, code := runHeadStdin(t, input, "-n", "4", "-n", "2")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout != "1\n2\n" {
		t.Errorf("stdout = %q, want %q", stdout, "1\n2\n")
	}
}

// ---------------------------------------------------------------------------
// 4. Negative Line Counts (-n -NUM)
// ---------------------------------------------------------------------------

func TestHead_NegativeLines_Stdin(t *testing.T) {
	input := "line 1\nline 2\nline 3\nline 4\nline 5\n"
	stdout, _, code := runHeadStdin(t, input, "-n", "-2")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	expected := "line 1\nline 2\nline 3\n"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestHead_NegativeLines_File(t *testing.T) {
	input := "A\nB\nC\nD\nE\nF\n"
	f := writeFile(t, "letters.txt", input)

	stdout, _, code := runHead(t, "-n", "-3", f)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	expected := "A\nB\nC\n"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestHead_NegativeLines_ExceedTotalLines(t *testing.T) {
	input := "1\n2\n3\n"
	stdout, _, code := runHeadStdin(t, input, "-n", "-5")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
}

func TestHead_NegativeLines_ExactMatch(t *testing.T) {
	input := "1\n2\n3\n"
	stdout, _, code := runHeadStdin(t, input, "-n", "-3")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
}

func TestHead_NegativeLines_CRLF(t *testing.T) {
	input := "1\r\n2\r\n3\r\n4\r\n5\r\n"
	stdout, _, code := runHeadStdin(t, input, "-n", "-2")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	expected := "1\r\n2\r\n3\r\n"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

// ---------------------------------------------------------------------------
// 5. Byte Mode (-c NUM)
// ---------------------------------------------------------------------------

func TestHead_Bytes_Zero(t *testing.T) {
	stdout, _, code := runHeadStdin(t, "hello world", "-c", "0")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
}

func TestHead_Bytes_Positive(t *testing.T) {
	input := "abcdefghijklmnopqrstuvwxyz"
	stdout, _, code := runHeadStdin(t, input, "-c", "5")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	expected := "abcde"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestHead_Bytes_ExplicitPlus(t *testing.T) {
	input := "abcdefghijklmnopqrstuvwxyz"
	stdout, _, code := runHeadStdin(t, input, "-c", "+4")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	expected := "abcd"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestHead_Bytes_ExceedsLength(t *testing.T) {
	input := "short"
	stdout, _, code := runHeadStdin(t, input, "-c", "100")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout != input {
		t.Errorf("stdout = %q, want %q", stdout, input)
	}
}

func TestHead_Bytes_BinaryData(t *testing.T) {
	raw := []byte{0x00, 0xFF, 0xFE, 0x01, 0x02, 0x03}
	stdout, _, code := runHeadStdin(t, string(raw), "-c", "3")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout != string(raw[:3]) {
		t.Errorf("binary stdout mismatch, got %v, want %v", []byte(stdout), raw[:3])
	}
}

func TestHead_Bytes_ModeInterplay(t *testing.T) {
	input := "1\n2\n3\n4\n5\n"
	// -c after -n -> byteMode wins
	stdout, _, code := runHeadStdin(t, input, "-n", "3", "-c", "4")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout != "1\n2\n" {
		t.Errorf("stdout = %q, want %q", stdout, "1\n2\n")
	}

	// -n after -c -> lineMode wins
	stdout, _, code = runHeadStdin(t, input, "-c", "4", "-n", "1")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout != "1\n" {
		t.Errorf("stdout = %q, want %q", stdout, "1\n")
	}
}

// ---------------------------------------------------------------------------
// 6. Negative Byte Counts (-c -NUM)
// ---------------------------------------------------------------------------

func TestHead_NegativeBytes_StdinStream(t *testing.T) {
	input := "The quick brown fox jumps over the lazy dog"
	// Total length = 43. Subtract 4 -> 39 bytes
	stdout, _, code := runHeadStdin(t, input, "-c", "-4")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	expected := input[:len(input)-4]
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestHead_NegativeBytes_SeekableFile(t *testing.T) {
	input := "0123456789ABCDEF"
	f := writeFile(t, "hex.txt", input)

	stdout, _, code := runHead(t, "-c", "-6", f)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	expected := "0123456789"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestHead_NegativeBytes_ExceedsTotal(t *testing.T) {
	input := "abcdef"
	stdout, _, code := runHeadStdin(t, input, "-c", "-10")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
}

func TestHead_NegativeBytes_ExactLength(t *testing.T) {
	input := "abcdef"
	stdout, _, code := runHeadStdin(t, input, "-c", "-6")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
}

// ---------------------------------------------------------------------------
// 7. Delimiter Mode (-z, --zero-terminated)
// ---------------------------------------------------------------------------

func TestHead_ZeroTerminated_Positive(t *testing.T) {
	input := "record 1\x00record 2\x00record 3\x00record 4\x00"
	stdout, _, code := runHeadStdin(t, input, "-z", "-n", "2")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	expected := "record 1\x00record 2\x00"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestHead_ZeroTerminated_WithEmbeddedNewlines(t *testing.T) {
	input := "record\n1\x00record\n2\x00record\n3\x00"
	stdout, _, code := runHeadStdin(t, input, "-z", "-n", "2")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	expected := "record\n1\x00record\n2\x00"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestHead_ZeroTerminated_Negative(t *testing.T) {
	input := "a\x00b\x00c\x00d\x00"
	stdout, _, code := runHeadStdin(t, input, "-z", "-n", "-1")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	expected := "a\x00b\x00c\x00"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

// ---------------------------------------------------------------------------
// 8. Headers and Multiple Files
// ---------------------------------------------------------------------------

func TestHead_Headers_SingleFileNoHeader(t *testing.T) {
	f := writeFile(t, "single.txt", "line1\nline2\n")
	stdout, _, code := runHead(t, f)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if strings.Contains(stdout, "==>") {
		t.Errorf("single file should not have header by default, got: %s", stdout)
	}
	expected := "line1\nline2\n"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestHead_Headers_MultipleFiles(t *testing.T) {
	f1 := writeFile(t, "f1.txt", "one\n")
	f2 := writeFile(t, "f2.txt", "two\n")
	f3 := writeFile(t, "f3.txt", "three\n")

	stdout, _, code := runHead(t, f1, f2, f3)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	expected := fmt.Sprintf("==> %s <==\none\n\n==> %s <==\ntwo\n\n==> %s <==\nthree\n", f1, f2, f3)
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestHead_Headers_StdinWithFile(t *testing.T) {
	f1 := writeFile(t, "first.txt", "content1\n")
	stdout, _, code := runHeadStdin(t, "piped\n", f1, "-")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	expected := fmt.Sprintf("==> %s <==\ncontent1\n\n==> standard input <==\npiped\n", f1)
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestHead_Headers_QuietFlag(t *testing.T) {
	f1 := writeFile(t, "q1.txt", "first\n")
	f2 := writeFile(t, "q2.txt", "second\n")

	stdout, _, code := runHead(t, "-q", f1, f2)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	expected := "first\nsecond\n"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}

	// Long flag --quiet
	stdout, _, code = runHead(t, "--quiet", f1, f2)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestHead_Headers_VerboseFlag(t *testing.T) {
	f := writeFile(t, "v1.txt", "data\n")

	// Single file with -v
	stdout, _, code := runHead(t, "-v", f)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	expected := fmt.Sprintf("==> %s <==\ndata\n", f)
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}

	// Stdin with -v
	stdout, _, code = runHeadStdin(t, "stdin data\n", "-v")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	expected = "==> standard input <==\nstdin data\n"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestHead_Headers_QuietAndVerbosePrecedence(t *testing.T) {
	f1 := writeFile(t, "p1.txt", "one\n")
	f2 := writeFile(t, "p2.txt", "two\n")

	// -q followed by -v: verbose wins
	stdout, _, code := runHead(t, "-q", "-v", f1, f2)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout, "==>") {
		t.Errorf("expected headers when -v follows -q, got: %s", stdout)
	}

	// -v followed by -q: quiet wins
	stdout, _, code = runHead(t, "-v", "-q", f1, f2)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if strings.Contains(stdout, "==>") {
		t.Errorf("expected no headers when -q follows -v, got: %s", stdout)
	}
}

// ---------------------------------------------------------------------------
// 9. Error Handling & Edge Cases
// ---------------------------------------------------------------------------

func TestHead_Errors_NonExistentFile(t *testing.T) {
	_, stderr, code := runHead(t, "nonexistent_file_12345.txt")
	if code != 1 {
		t.Errorf("expected exit code 1 for missing file, got %d", code)
	}
	if !strings.Contains(stderr, "cannot open 'nonexistent_file_12345.txt'") {
		t.Errorf("stderr missing expected error, got: %s", stderr)
	}
}

func TestHead_Errors_MultiFilePartialFailure(t *testing.T) {
	fValid := writeFile(t, "valid.txt", "valid content\n")
	fMissing := filepath.Join(t.TempDir(), "missing.txt")

	// Missing file first, then valid file
	stdout, stderr, code := runHead(t, fMissing, fValid)
	if code != 1 {
		t.Errorf("expected exit code 1 on partial failure, got %d", code)
	}
	if !strings.Contains(stderr, "cannot open '"+fMissing+"'") {
		t.Errorf("stderr missing error for missing file: %s", stderr)
	}
	if !strings.Contains(stdout, "valid content\n") {
		t.Errorf("stdout should contain valid file content, got: %s", stdout)
	}

	// Valid file first, then missing file
	stdout, stderr, code = runHead(t, fValid, fMissing)
	if code != 1 {
		t.Errorf("expected exit code 1 on partial failure, got %d", code)
	}
	if !strings.Contains(stderr, "cannot open '"+fMissing+"'") {
		t.Errorf("stderr missing error for missing file: %s", stderr)
	}
	if !strings.Contains(stdout, "valid content\n") {
		t.Errorf("stdout should contain valid file content, got: %s", stdout)
	}
}

func TestHead_Errors_UnknownFlag(t *testing.T) {
	_, stderr, code := runHeadStdin(t, "data", "--unknown-flag-xyz")
	if code != 1 {
		t.Errorf("expected exit code 1 for unknown flag, got %d", code)
	}
	if !strings.Contains(stderr, "unknown") && !strings.Contains(stderr, "unrecognized") {
		t.Errorf("stderr missing flag diagnostic: %s", stderr)
	}
}

func TestHead_BrokenPipe_Handling(t *testing.T) {
	cmd := newCommand()
	err := cmd.Execute([]string{"-n", "10"}, strings.NewReader("a\nb\nc\n"), &brokenPipeWriter{}, &bytes.Buffer{})
	if err != nil && !errors.Is(err, syscall.EPIPE) {
		t.Errorf("expected broken pipe to be handled gracefully, got error: %v", err)
	}
}

func TestHead_InStreaming_InfiniteInput(t *testing.T) {
	// Verify that reading from a stream with -n stops after requested lines
	r, w := io.Pipe()
	go func() {
		defer w.Close()
		for i := 0; i < 1000; i++ {
			fmt.Fprintf(w, "stream line %d\n", i)
		}
	}()

	stdout, stderr, err := executeMemory([]string{"-n", "3"}, r)
	if err != nil {
		t.Fatalf("unexpected error: %v, stderr: %s", err, stderr)
	}
	expected := "stream line 0\nstream line 1\nstream line 2\n"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

// ---------------------------------------------------------------------------
// 10. Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkHead_First10Lines(b *testing.B) {
	var sb strings.Builder
	for range 1000 {
		sb.WriteString("benchmark line of text for testing throughput\n")
	}
	content := sb.String()

	for b.Loop() {
		var out bytes.Buffer
		_ = copyFirstLines(&out, strings.NewReader(content), '\n', 10)
	}
}

func BenchmarkHead_NegativeLines(b *testing.B) {
	var sb strings.Builder
	for range 1000 {
		sb.WriteString("benchmark line of text for testing throughput\n")
	}
	content := sb.String()

	for b.Loop() {
		var out bytes.Buffer
		_ = copyAllExceptLastLines(&out, strings.NewReader(content), '\n', 10)
	}
}

func BenchmarkHead_FirstBytes(b *testing.B) {
	data := bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz012345"), 1000)

	for b.Loop() {
		var out bytes.Buffer
		_ = copyFirstBytes(&out, bytes.NewReader(data), 1024)
	}
}

func BenchmarkHead_NegativeBytes(b *testing.B) {
	data := bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz012345"), 1000)

	for b.Loop() {
		var out bytes.Buffer
		_ = copyAllExceptLastBytes(&out, bytes.NewReader(data), 1024)
	}
}
