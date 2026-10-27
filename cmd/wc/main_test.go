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
	tmpDir, err := os.MkdirTemp("", "wc-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	name := "wc-test"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	bin := filepath.Join(tmpDir, name)
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build wc binary: %v\n%s\n", err, out)
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

func runWC(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	return runWCStdin(t, "", args...)
}

func runWCStdin(t *testing.T, stdin string, args ...string) (string, string, int) {
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
			t.Fatalf("failed to run wc: %v", err)
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

type errReader struct {
	err error
}

func (e *errReader) Read(_ []byte) (int, error) {
	return 0, e.err
}

// ---------------------------------------------------------------------------
// 1. Unit Tests: Metadata, Flags, countReader, & Formatting
// ---------------------------------------------------------------------------

func TestUnit_HelpAndVersion(t *testing.T) {
	help := newCommand().Help()
	snippets := []string{
		"Usage:",
		"wc [OPTION]... [FILE]...",
		"wc [OPTION]... --files0-from=F",
		"Description:",
		"Print newline, word, and byte counts for each FILE",
		"-c, --bytes",
		"-m, --chars",
		"-l, --lines",
		"--files0-from=F",
		"-L, --max-line-length",
		"-w, --words",
		"--help",
		"--version",
	}
	for _, snip := range snippets {
		if !strings.Contains(help, snip) {
			t.Errorf("help missing snippet %q\nFull help:\n%s", snip, help)
		}
	}

	out, errOut, err := executeMemory([]string{"--version"}, nil)
	if err != nil || errOut != "" {
		t.Fatalf("unexpected error on --version: err=%v, stderr=%q", err, errOut)
	}
	if !strings.Contains(out, "wc (coreutils-go)") {
		t.Errorf("unexpected version output: %q", out)
	}
}

func TestUnit_CountReader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected wcCounts
	}{
		{
			name:     "empty",
			input:    "",
			expected: wcCounts{lines: 0, words: 0, chars: 0, bytes: 0, maxLineLen: 0},
		},
		{
			name:     "single line with newline",
			input:    "hello world\n",
			expected: wcCounts{lines: 1, words: 2, chars: 12, bytes: 12, maxLineLen: 11},
		},
		{
			name:     "single line without trailing newline",
			input:    "hello world",
			expected: wcCounts{lines: 0, words: 2, chars: 11, bytes: 11, maxLineLen: 11},
		},
		{
			name:     "multiple whitespace types",
			input:    "  one\ttwo\nthree\r\nfour\ffive\vsix  ",
			expected: wcCounts{lines: 2, words: 6, chars: 32, bytes: 32, maxLineLen: 13},
		},
		{
			name:     "utf8 multibyte characters",
			input:    "Café ☕\n",
			expected: wcCounts{lines: 1, words: 2, chars: 7, bytes: 10, maxLineLen: 6},
		},
		{
			name:     "tab expansion at start of line",
			input:    "\tabc\n",
			expected: wcCounts{lines: 1, words: 1, chars: 5, bytes: 5, maxLineLen: 11}, // 8 + 3
		},
		{
			name:     "tab expansion mid line",
			input:    "ab\tc\n",
			expected: wcCounts{lines: 1, words: 2, chars: 5, bytes: 5, maxLineLen: 9}, // 2 + (8-2) + 1 = 9
		},
		{
			name:     "carriage return resets line length",
			input:    "12345\rabc\n",
			expected: wcCounts{lines: 1, words: 2, chars: 10, bytes: 10, maxLineLen: 5},
		},
		{
			name:     "non-printable control character ignored in maxLineLen",
			input:    "ab\x07cd\n",
			expected: wcCounts{lines: 1, words: 1, chars: 6, bytes: 6, maxLineLen: 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := countReader(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("countReader(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("countReader(%q) = %+v, want %+v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestUnit_CountReader_Error(t *testing.T) {
	boom := errors.New("disk read failure")
	_, err := countReader(&errReader{err: boom})
	if !errors.Is(err, boom) {
		t.Errorf("expected %v, got %v", boom, err)
	}
}

func TestUnit_ExecuteMemory_Direct(t *testing.T) {
	t.Run("default stdin", func(t *testing.T) {
		out, errOut, err := executeMemory([]string{}, strings.NewReader("The quick brown fox jumps over the lazy dog\n"))
		if err != nil || errOut != "" {
			t.Fatalf("err=%v, errOut=%q", err, errOut)
		}
		if out != "      1       9      44\n" {
			t.Errorf("out = %q, want %q", out, "      1       9      44\n")
		}
	})

	t.Run("all flags single file", func(t *testing.T) {
		f := writeFile(t, "direct.txt", "Café ☕\n")
		out, errOut, err := executeMemory([]string{"-l", "-w", "-m", "-c", "-L", f}, nil)
		if err != nil || errOut != "" {
			t.Fatalf("err=%v, errOut=%q", err, errOut)
		}
		expected := fmt.Sprintf("      1       2       7      10       6 %s\n", f)
		if out != expected {
			t.Errorf("out = %q, want %q", out, expected)
		}
	})

	t.Run("files0-from stdin", func(t *testing.T) {
		f1 := writeFile(t, "f1.txt", "a b\n")
		f2 := writeFile(t, "f2.txt", "c d e\n")
		stream := f1 + "\x00" + f2 + "\x00"
		out, errOut, err := executeMemory([]string{"-w", "--files0-from=-"}, strings.NewReader(stream))
		if err != nil || errOut != "" {
			t.Fatalf("err=%v, errOut=%q", err, errOut)
		}
		expected := fmt.Sprintf("2 %s\n3 %s\n5 total\n", f1, f2)
		if out != expected {
			t.Errorf("out = %q, want %q", out, expected)
		}
	})

	t.Run("error paths in-process", func(t *testing.T) {
		// Unknown flag
		if _, errOut, err := executeMemory([]string{"--bad-flag"}, nil); err == nil || !strings.Contains(errOut, "Try 'wc --help'") {
			t.Errorf("expected unknown flag error, got err=%v, errOut=%q", err, errOut)
		}
		// Empty files0-from
		if _, errOut, err := executeMemory([]string{"--files0-from="}, nil); err == nil || !strings.Contains(errOut, "cannot open ''") {
			t.Errorf("expected empty files0-from error, got err=%v, errOut=%q", err, errOut)
		}
		// Missing files0-from file
		if _, errOut, err := executeMemory([]string{"--files0-from=missing_list.txt"}, nil); err == nil || !strings.Contains(errOut, "cannot open 'missing_list.txt'") {
			t.Errorf("expected missing files0-from error, got err=%v, errOut=%q", err, errOut)
		}
		// Read error on files0-from stdin
		if _, _, err := executeMemory([]string{"--files0-from=-"}, &errReader{err: errors.New("read failed")}); err == nil {
			t.Errorf("expected error when reading files0-from stream fails")
		}
		// Zero-length filename and missing file in-process
		if _, errOut, err := executeMemory([]string{"--files0-from=-"}, strings.NewReader("\x00missing_target.txt\x00")); err == nil || !strings.Contains(errOut, "invalid zero-length file name") {
			t.Errorf("expected zero-length filename error, got err=%v, errOut=%q", err, errOut)
		}
		// Read error on stdin operand
		if _, errOut, err := executeMemory([]string{"-"}, &errReader{err: errors.New("stdin read error")}); err == nil || !strings.Contains(errOut, "stdin read error") {
			t.Errorf("expected stdin read error, got err=%v, errOut=%q", err, errOut)
		}
	})
}

// ---------------------------------------------------------------------------
// 2. Default Counts (-l -w -c)
// ---------------------------------------------------------------------------

func TestWC_Default_Stdin(t *testing.T) {
	input := "The quick brown fox jumps over the lazy dog\n"
	stdout, stderr, code := runWCStdin(t, input)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}
	expected := "      1       9      44\n"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestWC_Default_EmptyInput(t *testing.T) {
	stdout, stderr, code := runWCStdin(t, "")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}
	expected := "      0       0       0\n"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestWC_Default_NoTrailingNewline(t *testing.T) {
	input := "line without newline"
	stdout, stderr, code := runWCStdin(t, input)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}
	expected := "      0       3      20\n"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestWC_Default_SingleFile(t *testing.T) {
	f := writeFile(t, "sample.txt", "alpha beta\ngamma\n")
	stdout, stderr, code := runWC(t, f)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}
	expected := fmt.Sprintf("      2       3      17 %s\n", f)
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

// ---------------------------------------------------------------------------
// 3. Individual Count Flags (-l, -w, -c, -m, -L)
// ---------------------------------------------------------------------------

func TestWC_LinesOnly(t *testing.T) {
	input := "line 1\nline 2\nline 3\n"
	stdout, stderr, code := runWCStdin(t, input, "-l")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}
	if stdout != "3\n" {
		t.Errorf("stdout = %q, want %q", stdout, "3\n")
	}

	// Long option --lines with file
	f := writeFile(t, "lines.txt", input)
	stdout, stderr, code = runWC(t, "--lines", f)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}
	expected := fmt.Sprintf("3 %s\n", f)
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestWC_WordsOnly(t *testing.T) {
	input := "   one   two\tthree\n\nfour\r\nfive   "
	stdout, stderr, code := runWCStdin(t, input, "-w")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}
	if stdout != "5\n" {
		t.Errorf("stdout = %q, want %q", stdout, "5\n")
	}

	stdout, _, code = runWCStdin(t, input, "--words")
	if code != 0 || stdout != "5\n" {
		t.Errorf("stdout = %q, code = %d", stdout, code)
	}
}

func TestWC_BytesVsChars_UTF8(t *testing.T) {
	// "Café ☕" -> 'C'(1) 'a'(1) 'f'(1) 'é'(2) ' '(1) '☕'(3) = 6 chars, 9 bytes?
	// Wait: 'C'(1) + 'a'(1) + 'f'(1) + 'é'(2) + ' '(1) + '☕'(3) = 9 bytes, or with variation selector?
	// Let's test exact len(input) and utf8.RuneCountInString(input)
	input := "Café ☕"
	stdoutBytes, _, code := runWCStdin(t, input, "-c")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdoutBytes != fmt.Sprintf("%d\n", len(input)) {
		t.Errorf("bytes = %q, want %d", stdoutBytes, len(input))
	}

	stdoutChars, _, code := runWCStdin(t, input, "-m")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdoutChars != "6\n" {
		t.Errorf("chars = %q, want %q", stdoutChars, "6\n")
	}

	// Long flags --bytes and --chars
	stdoutBytesLong, _, _ := runWCStdin(t, input, "--bytes")
	stdoutCharsLong, _, _ := runWCStdin(t, input, "--chars")
	if stdoutBytesLong != stdoutBytes || stdoutCharsLong != stdoutChars {
		t.Errorf("long flags mismatch: bytes=%q, chars=%q", stdoutBytesLong, stdoutCharsLong)
	}
}

func TestWC_MaxLineLength(t *testing.T) {
	input := "short\nthis is the longest line in the input\nmedium line\n"
	stdout, stderr, code := runWCStdin(t, input, "-L")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}
	// len("this is the longest line in the input") == 37
	if stdout != "37\n" {
		t.Errorf("stdout = %q, want %q", stdout, "37\n")
	}

	// Long option --max-line-length with tabs
	// "\t1234" -> 8 + 4 = 12
	stdout, _, code = runWCStdin(t, "\t1234\n1234567890\n", "--max-line-length")
	if code != 0 || stdout != "12\n" {
		t.Errorf("stdout = %q, want %q (code=%d)", stdout, "12\n", code)
	}
}

// ---------------------------------------------------------------------------
// 4. Canonical Column Ordering & Flag Combinations
// ---------------------------------------------------------------------------

func TestWC_CanonicalColumnOrdering(t *testing.T) {
	// Canonical order: lines, words, chars, bytes, max-line-length
	// Input: "Café\n" -> 1 line, 1 word, 5 chars, 6 bytes, 4 max-line-length
	input := "Café\n"

	// Pass flags in reverse order: -L -c -m -w -l
	stdout1, _, code1 := runWCStdin(t, input, "-L", "-c", "-m", "-w", "-l")
	// Pass flags in canonical order: -l -w -m -c -L
	stdout2, _, code2 := runWCStdin(t, input, "-l", "-w", "-m", "-c", "-L")
	// Pass clustered flags: -Lwclm
	stdout3, _, code3 := runWCStdin(t, input, "-Lwclm")

	if code1 != 0 || code2 != 0 || code3 != 0 {
		t.Fatalf("non-zero exit codes: %d, %d, %d", code1, code2, code3)
	}
	expected := "      1       1       5       6       4\n"
	if stdout1 != expected || stdout2 != expected || stdout3 != expected {
		t.Errorf("canonical order mismatch:\nstdout1=%q\nstdout2=%q\nstdout3=%q\nwant   =%q", stdout1, stdout2, stdout3, expected)
	}
}

// ---------------------------------------------------------------------------
// 5. Multiple Files & Total Row
// ---------------------------------------------------------------------------

func TestWC_MultipleFiles_TotalRow(t *testing.T) {
	f1 := writeFile(t, "file1.txt", "one two\nthree\n")               // 2 lines, 3 words, 14 bytes, maxLineLen 7
	f2 := writeFile(t, "file2.txt", "four five six seven\n")          // 1 line, 4 words, 20 bytes, maxLineLen 19
	f3 := writeFile(t, "file3.txt", "eight\nnine\nten\neleven\n")     // 4 lines, 4 words, 22 bytes, maxLineLen 6

	stdout, stderr, code := runWC(t, f1, f2, f3)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}

	// Total: 7 lines, 11 words, 56 bytes -> maxVal is 56 (2 digits)
	expected := fmt.Sprintf(" 2  3 14 %s\n 1  4 20 %s\n 4  4 22 %s\n 7 11 56 total\n", f1, f2, f3)
	if stdout != expected {
		t.Errorf("stdout =\n%q\nwant =\n%q", stdout, expected)
	}
}

func TestWC_MultipleFiles_MaxLineLengthIsMaxNotSum(t *testing.T) {
	f1 := writeFile(t, "max1.txt", "12345\n")          // maxLineLen = 5
	f2 := writeFile(t, "max2.txt", "123456789012\n")   // maxLineLen = 12
	f3 := writeFile(t, "max3.txt", "12345678\n")       // maxLineLen = 8

	stdout, stderr, code := runWC(t, "-L", f1, f2, f3)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}

	// Total max-line-length must be 12 (the maximum), not 25 (the sum)
	expected := fmt.Sprintf(" 5 %s\n12 %s\n 8 %s\n12 total\n", f1, f2, f3)
	if stdout != expected {
		t.Errorf("stdout =\n%q\nwant =\n%q", stdout, expected)
	}
}

// ---------------------------------------------------------------------------
// 6. --files0-from=F
// ---------------------------------------------------------------------------

func TestWC_Files0From_File(t *testing.T) {
	f1 := writeFile(t, "a.txt", "hello\n")
	f2 := writeFile(t, "b.txt", "world again\n")
	listFile := writeFile(t, "files.list", f1+"\x00"+f2+"\x00")

	stdout, stderr, code := runWC(t, "-l", "-w", "--files0-from="+listFile)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}

	expected := fmt.Sprintf("1 1 %s\n1 2 %s\n2 3 total\n", f1, f2)
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestWC_Files0From_Stdin(t *testing.T) {
	f1 := writeFile(t, "s1.txt", "line1\nline2\n")
	f2 := writeFile(t, "s2.txt", "line3\n")
	stdinList := f1 + "\x00" + f2

	stdout, stderr, code := runWCStdin(t, stdinList, "-l", "--files0-from=-")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}

	expected := fmt.Sprintf("2 %s\n1 %s\n3 total\n", f1, f2)
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestWC_Files0From_EmptyList(t *testing.T) {
	listFile := writeFile(t, "empty.list", "")
	stdout, stderr, code := runWC(t, "--files0-from="+listFile)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
}

func TestWC_Files0From_ZeroLengthFilename(t *testing.T) {
	f1 := writeFile(t, "valid.txt", "ok\n")
	// Contains an empty NUL-terminated entry between two NULs
	listFile := writeFile(t, "bad.list", f1+"\x00\x00")

	stdout, stderr, code := runWC(t, "-l", "--files0-from="+listFile)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr, "invalid zero-length file name") {
		t.Errorf("expected zero-length file name error in stderr, got: %s", stderr)
	}
	if !strings.Contains(stdout, f1) {
		t.Errorf("expected valid file to still be processed, got stdout: %q", stdout)
	}
}

func TestWC_Files0From_Errors(t *testing.T) {
	// Empty argument --files0-from=
	_, stderr, code := runWC(t, "--files0-from=")
	if code != 1 {
		t.Errorf("expected exit code 1 for empty --files0-from=, got %d", code)
	}
	if !strings.Contains(stderr, "cannot open '' for reading") {
		t.Errorf("unexpected stderr: %s", stderr)
	}

	// Non-existent list file
	_, stderr, code = runWC(t, "--files0-from=nonexistent_list_999.txt")
	if code != 1 {
		t.Errorf("expected exit code 1 for missing files0-from file, got %d", code)
	}
	if !strings.Contains(stderr, "cannot open 'nonexistent_list_999.txt' for reading") {
		t.Errorf("unexpected stderr: %s", stderr)
	}
}

// ---------------------------------------------------------------------------
// 7. Error Handling & Edge Cases
// ---------------------------------------------------------------------------

func TestWC_Errors_NonExistentFile(t *testing.T) {
	_, stderr, code := runWC(t, "nonexistent_file_12345.txt")
	if code != 1 {
		t.Errorf("expected exit code 1 for missing file, got %d", code)
	}
	if !strings.Contains(stderr, "wc: nonexistent_file_12345.txt:") {
		t.Errorf("stderr missing expected error, got: %s", stderr)
	}
}

func TestWC_Errors_MultiFilePartialFailure(t *testing.T) {
	fValid := writeFile(t, "valid.txt", "hello world\n")
	fMissing := filepath.Join(t.TempDir(), "missing.txt")

	stdout, stderr, code := runWC(t, fMissing, fValid)
	if code != 1 {
		t.Errorf("expected exit code 1 on partial failure, got %d", code)
	}
	if !strings.Contains(stderr, fMissing) {
		t.Errorf("stderr missing error for missing file: %s", stderr)
	}
	if !strings.Contains(stdout, fValid) || !strings.Contains(stdout, "total") {
		t.Errorf("stdout should contain valid file counts and total, got: %s", stdout)
	}
}

func TestWC_Errors_UnknownFlag(t *testing.T) {
	_, stderr, code := runWCStdin(t, "data", "--unknown-flag-xyz")
	if code != 1 {
		t.Errorf("expected exit code 1 for unknown flag, got %d", code)
	}
	if !strings.Contains(stderr, "unrecognized option '--unknown-flag-xyz'") {
		t.Errorf("stderr missing flag diagnostic: %s", stderr)
	}
	if !strings.Contains(stderr, "Try 'wc --help' for more information.") {
		t.Errorf("stderr missing help hint: %s", stderr)
	}
}

func TestWC_BrokenPipe_Handling(t *testing.T) {
	f1 := writeFile(t, "bp1.txt", "a\nb\n")
	f2 := writeFile(t, "bp2.txt", "c\nd\n")
	cmd := newCommand()
	err := cmd.Execute([]string{f1, f2}, strings.NewReader(""), &brokenPipeWriter{}, &bytes.Buffer{})
	if err != nil && !errors.Is(err, syscall.EPIPE) {
		t.Errorf("expected broken pipe to be handled gracefully, got error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// 8. Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkWC_CountReader_ASCII(b *testing.B) {
	var sb strings.Builder
	for range 1000 {
		sb.WriteString("The quick brown fox jumps over the lazy dog 0123456789\n")
	}
	content := sb.String()

	for b.Loop() {
		_, _ = countReader(strings.NewReader(content))
	}
}

func BenchmarkWC_CountReader_UTF8(b *testing.B) {
	var sb strings.Builder
	for range 1000 {
		sb.WriteString("Café ☕ naïve résumé 日本語テキスト\twith\ttabs\n")
	}
	content := sb.String()

	for b.Loop() {
		_, _ = countReader(strings.NewReader(content))
	}
}

func BenchmarkWC_ExecuteMemory_AllFlags(b *testing.B) {
	var sb strings.Builder
	for range 1000 {
		sb.WriteString("benchmark line of text for testing wc throughput\n")
	}
	content := sb.String()

	for b.Loop() {
		_, _, _ = executeMemory([]string{"-l", "-w", "-m", "-c", "-L"}, strings.NewReader(content))
	}
}
