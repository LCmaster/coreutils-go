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
	"time"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "cat-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	name := "cat-test"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	bin := filepath.Join(tmpDir, name)
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build cat binary: %v\n%s\n", err, out)
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

func runCat(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	return runCatStdin(t, "", args...)
}

func runCatStdin(t *testing.T, stdin string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("failed to run cat: %v", err)
		}
		code = exitErr.ExitCode()
	}
	return stdout.String(), stderr.String(), code
}

func writeFile(t *testing.T, name string, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
	return path
}

// ---------------------------------------------------------------------------
// Unit tests
// ---------------------------------------------------------------------------

func TestUnit_PrintHelp(t *testing.T) {
	help := newCommand().Help()
	for _, snippet := range []string{
		"Usage:",
		"cat [OPTION]... [FILE]...",
		"Description:",
		"Concatenate FILE(s) to standard output.",
		"With no FILE, or when FILE is -, read standard input.",
		"-A, --show-all",
		"-b, --number-nonblank",
		"-e",
		"-E, --show-ends",
		"-n, --number",
		"-s, --squeeze-blank",
		"-t",
		"-T, --show-tabs",
		"-u",
		"-v, --show-nonprinting",
		"--help",
		"--version",
		"Examples:",
	} {
		if !strings.Contains(help, snippet) {
			t.Errorf("cmd.Help() missing snippet %q", snippet)
		}
	}
	if !strings.HasSuffix(help, "\n") {
		t.Error("cmd.Help() should end with a newline")
	}
}

func TestUnit_PrintVersion(t *testing.T) {
	ver := newCommand().Version()
	if ver != "cat (coreutils-go) 0.1.0" {
		t.Errorf("cmd.Version() = %q, want %q", ver, "cat (coreutils-go) 0.1.0")
	}
}

func TestUnit_FormatTabs(t *testing.T) {
	input := []byte("a\tb\t\n")
	want := []byte("a^Ib^I\n")
	got := formatTabs(input)
	if !bytes.Equal(got, want) {
		t.Errorf("formatTabs(%q) = %q, want %q", input, got, want)
	}
}

func TestUnit_FormatEnds(t *testing.T) {
	input := []byte("hello\n")
	want := []byte("hello$\n")
	got := formatEnds(input)
	if !bytes.Equal(got, want) {
		t.Errorf("formatEnds(%q) = %q, want %q", input, got, want)
	}
}

func TestUnit_ProcessStream(t *testing.T) {
	opts := CatOptions{
		Number:   true,
		ShowEnds: true,
	}
	pipeline := newFormatPipeline(opts)
	state := StreamState{LineNumber: 1}

	input := strings.NewReader("line1\nline2\n")
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	if err := processStream(w, input, opts, &state, pipeline); err != nil {
		t.Fatalf("processStream error: %v", err)
	}
	_ = w.Flush()

	want := "     1\tline1$\n     2\tline2$\n"
	if got := buf.String(); got != want {
		t.Errorf("processStream output = %q, want %q", got, want)
	}
	if state.LineNumber != 3 {
		t.Errorf("state.LineNumber = %d, want 3", state.LineNumber)
	}
}

func TestUnit_Flags(t *testing.T) {
	// Test -b overrides -n
	var out bytes.Buffer
	err := newCommand().Execute([]string{"-b", "-n"}, strings.NewReader("line1\n\nline2\n"), &out, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	expected := "     1\tline1\n\n     2\tline2\n"
	if out.String() != expected {
		t.Errorf("got %q, want %q", out.String(), expected)
	}

	// Test -e shorthand (ShowNonPrinting and ShowEnds)
	out.Reset()
	err = newCommand().Execute([]string{"-e"}, strings.NewReader("hello\n"), &out, io.Discard)
	if err != nil || out.String() != "hello$\n" {
		t.Errorf("-e got %q", out.String())
	}

	// Test -t shorthand (ShowNonPrinting and ShowTabs)
	out.Reset()
	err = newCommand().Execute([]string{"-t"}, strings.NewReader("a\tb\n"), &out, io.Discard)
	if err != nil || out.String() != "a^Ib\n" {
		t.Errorf("-t got %q", out.String())
	}

	// Test invalid option -h
	var errBuf bytes.Buffer
	err = newCommand().Execute([]string{"-h"}, nil, io.Discard, &errBuf)
	if err == nil || !strings.Contains(errBuf.String(), "invalid option -- 'h'") {
		t.Errorf("expected invalid option -h, got err=%v, stderr=%q", err, errBuf.String())
	}

	// Test unrecognized long option --unknown
	errBuf.Reset()
	err = newCommand().Execute([]string{"--unknown"}, nil, io.Discard, &errBuf)
	if err == nil || !strings.Contains(errBuf.String(), "unrecognized option '--unknown'") {
		t.Errorf("expected unrecognized option --unknown, got err=%v, stderr=%q", err, errBuf.String())
	}
}

func TestUnit_Run(t *testing.T) {
	// Test Help
	var outHelp, errHelp bytes.Buffer
	err := newCommand().Execute([]string{"--help"}, nil, &outHelp, &errHelp)
	if err != nil || !strings.Contains(outHelp.String(), "Usage:") || errHelp.Len() > 0 {
		t.Errorf("Execute(--help) err=%v, stdout=%q, stderr=%q", err, outHelp.String(), errHelp.String())
	}

	// Test Version
	var outVer, errVer bytes.Buffer
	err = newCommand().Execute([]string{"--version"}, nil, &outVer, &errVer)
	if err != nil || outVer.String() != "cat (coreutils-go) 0.1.0\n" || errVer.Len() > 0 {
		t.Errorf("Execute(--version) err=%v, stdout=%q, stderr=%q", err, outVer.String(), errVer.String())
	}

	// Test Invalid Option
	var outInv, errInv bytes.Buffer
	err = newCommand().Execute([]string{"--unknown-flag"}, nil, &outInv, &errInv)
	if err == nil || !strings.Contains(errInv.String(), "unrecognized option '--unknown-flag'") {
		t.Errorf("Execute(--unknown-flag) err=%v, stderr=%q", err, errInv.String())
	}

	// Test Success Run
	var outRun, errRun bytes.Buffer
	err = newCommand().Execute([]string{}, strings.NewReader("hello cat\n"), &outRun, &errRun)
	if err != nil || outRun.String() != "hello cat\n" || errRun.Len() > 0 {
		t.Errorf("Execute() err=%v, stdout=%q, stderr=%q", err, outRun.String(), errRun.String())
	}
}

// ---------------------------------------------------------------------------
// --help / --version
// ---------------------------------------------------------------------------

func TestE2E_Help(t *testing.T) {
	stdout, stderr, code := runCat(t, "--help")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr %q)", code, stderr)
	}
	expectedHelp := newCommand().Help()
	if stdout != expectedHelp {
		t.Errorf("--help output mismatch:\n got: %q\nwant: %q", stdout, expectedHelp)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestE2E_HelpIgnoresStdin(t *testing.T) {
	stdout, _, code := runCatStdin(t, "should not be echoed\n", "--help")
	if code != 0 || strings.Contains(stdout, "should not be echoed") {
		t.Errorf("--help must not read stdin; code=%d stdout=%q", code, stdout)
	}
}

func TestE2E_Version(t *testing.T) {
	stdout, stderr, code := runCat(t, "--version")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr %q)", code, stderr)
	}
	if stdout != "cat (coreutils-go) 0.1.0\n" {
		t.Errorf("--version = %q", stdout)
	}
}

// ---------------------------------------------------------------------------
// Basic concatenation
// ---------------------------------------------------------------------------

func TestE2E_SingleFile(t *testing.T) {
	f := writeFile(t, "a.txt", "hello\nworld\n")
	stdout, stderr, code := runCat(t, f)
	if code != 0 || stderr != "" || stdout != "hello\nworld\n" {
		t.Errorf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestE2E_MultipleFiles(t *testing.T) {
	a := writeFile(t, "a.txt", "one\n")
	b := writeFile(t, "b.txt", "two\n")
	c := writeFile(t, "c.txt", "three\n")
	stdout, _, code := runCat(t, a, b, c)
	if code != 0 || stdout != "one\ntwo\nthree\n" {
		t.Errorf("code=%d stdout=%q", code, stdout)
	}
}

func TestE2E_SameFileTwice(t *testing.T) {
	a := writeFile(t, "a.txt", "x\n")
	stdout, _, code := runCat(t, a, a)
	if code != 0 || stdout != "x\nx\n" {
		t.Errorf("code=%d stdout=%q", code, stdout)
	}
}

func TestE2E_NoTrailingNewlinePreserved(t *testing.T) {
	f := writeFile(t, "a.txt", "no newline")
	stdout, _, _ := runCat(t, f)
	if stdout != "no newline" {
		t.Errorf("stdout=%q", stdout)
	}
}

func TestE2E_FilesWithoutTrailingNewlineAreJoined(t *testing.T) {
	a := writeFile(t, "a.txt", "foo")
	b := writeFile(t, "b.txt", "bar\n")
	stdout, _, _ := runCat(t, a, b)
	if stdout != "foobar\n" {
		t.Errorf("stdout=%q, want %q", stdout, "foobar\n")
	}
}

func TestE2E_EmptyFile(t *testing.T) {
	f := writeFile(t, "empty.txt", "")
	stdout, stderr, code := runCat(t, f)
	if code != 0 || stdout != "" || stderr != "" {
		t.Errorf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestE2E_EmptyBetweenFiles(t *testing.T) {
	a := writeFile(t, "a.txt", "a\n")
	e := writeFile(t, "e.txt", "")
	b := writeFile(t, "b.txt", "b\n")
	stdout, _, _ := runCat(t, a, e, b)
	if stdout != "a\nb\n" {
		t.Errorf("stdout=%q", stdout)
	}
}

func TestE2E_BinaryPassThrough(t *testing.T) {
	var data []byte
	for i := 0; i < 256; i++ {
		data = append(data, byte(i))
	}
	f := writeFile(t, "bin.dat", string(data))
	stdout, _, code := runCat(t, f)
	if code != 0 || stdout != string(data) {
		t.Errorf("binary data was altered (code=%d, len=%d)", code, len(stdout))
	}
}

func TestE2E_CRLFPreserved(t *testing.T) {
	f := writeFile(t, "crlf.txt", "a\r\nb\r\n")
	stdout, _, _ := runCat(t, f)
	if stdout != "a\r\nb\r\n" {
		t.Errorf("stdout=%q", stdout)
	}
}

func TestE2E_LargeFile(t *testing.T) {
	// Larger than the internal 64 KB buffer and not a multiple of it.
	line := strings.Repeat("0123456789abcdef", 4) + "\n" // 65 bytes
	content := strings.Repeat(line, 5000)                // ~325 KB
	f := writeFile(t, "large.txt", content)
	stdout, _, code := runCat(t, f)
	if code != 0 || stdout != content {
		t.Errorf("large file mismatch (code=%d, got %d bytes, want %d)", code, len(stdout), len(content))
	}
}

func TestE2E_VeryLongLine(t *testing.T) {
	content := strings.Repeat("x", 200*1024) + "\n"
	f := writeFile(t, "long.txt", content)
	stdout, _, _ := runCat(t, f, "-n")
	if want := fmt.Sprintf("%6d\t", 1) + content; stdout != want {
		t.Errorf("long line mismatch (got %d bytes, want %d)", len(stdout), len(want))
	}
}

func TestE2E_UnicodePassThrough(t *testing.T) {
	content := "héllo wörld ✓ 日本語\n"
	f := writeFile(t, "u.txt", content)
	stdout, _, _ := runCat(t, f)
	if stdout != content {
		t.Errorf("stdout=%q", stdout)
	}
}

func TestE2E_FilenameWithSpaces(t *testing.T) {
	f := writeFile(t, "file with spaces.txt", "ok\n")
	stdout, _, code := runCat(t, f)
	if code != 0 || stdout != "ok\n" {
		t.Errorf("code=%d stdout=%q", code, stdout)
	}
}

// ---------------------------------------------------------------------------
// Standard input
// ---------------------------------------------------------------------------

func TestE2E_StdinNoArgs(t *testing.T) {
	stdout, _, code := runCatStdin(t, "from stdin\n")
	if code != 0 || stdout != "from stdin\n" {
		t.Errorf("code=%d stdout=%q", code, stdout)
	}
}

func TestE2E_StdinDash(t *testing.T) {
	stdout, _, code := runCatStdin(t, "from stdin\n", "-")
	if code != 0 || stdout != "from stdin\n" {
		t.Errorf("code=%d stdout=%q", code, stdout)
	}
}

func TestE2E_StdinBetweenFiles(t *testing.T) {
	a := writeFile(t, "a.txt", "f\n")
	b := writeFile(t, "b.txt", "g\n")
	stdout, _, code := runCatStdin(t, "-in-\n", a, "-", b)
	if code != 0 || stdout != "f\n-in-\ng\n" {
		t.Errorf("code=%d stdout=%q", code, stdout)
	}
}

func TestE2E_StdinEmpty(t *testing.T) {
	stdout, _, code := runCatStdin(t, "")
	if code != 0 || stdout != "" {
		t.Errorf("code=%d stdout=%q", code, stdout)
	}
}

func TestE2E_StdinLarge(t *testing.T) {
	content := strings.Repeat("stdin line\n", 40000)
	stdout, _, code := runCatStdin(t, content)
	if code != 0 || stdout != content {
		t.Errorf("code=%d, got %d bytes, want %d", code, len(stdout), len(content))
	}
}

func TestE2E_StdinWithNumbering(t *testing.T) {
	stdout, _, _ := runCatStdin(t, "a\nb\n", "-n")
	if stdout != "     1\ta\n     2\tb\n" {
		t.Errorf("stdout=%q", stdout)
	}
}

// ---------------------------------------------------------------------------
// -n / --number
// ---------------------------------------------------------------------------

func TestE2E_Number(t *testing.T) {
	for _, flag := range []string{"-n", "--number"} {
		t.Run(flag, func(t *testing.T) {
			f := writeFile(t, "a.txt", "a\n\nb\n")
			stdout, _, code := runCat(t, flag, f)
			want := "     1\ta\n     2\t\n     3\tb\n"
			if code != 0 || stdout != want {
				t.Errorf("code=%d stdout=%q want=%q", code, stdout, want)
			}
		})
	}
}

func TestE2E_NumberContinuesAcrossFiles(t *testing.T) {
	a := writeFile(t, "a.txt", "a\nb\n")
	b := writeFile(t, "b.txt", "c\n")
	stdout, _, _ := runCat(t, "-n", a, b)
	if want := "     1\ta\n     2\tb\n     3\tc\n"; stdout != want {
		t.Errorf("stdout=%q want=%q", stdout, want)
	}
}

func TestE2E_NumberNoTrailingNewline(t *testing.T) {
	f := writeFile(t, "a.txt", "a\nb")
	stdout, _, _ := runCat(t, "-n", f)
	if want := "     1\ta\n     2\tb"; stdout != want {
		t.Errorf("stdout=%q want=%q", stdout, want)
	}
}

func TestE2E_NumberWidthGrows(t *testing.T) {
	content := strings.Repeat("x\n", 1000000/100) // 10000 lines
	f := writeFile(t, "a.txt", content)
	stdout, _, _ := runCat(t, "-n", f)
	lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
	if len(lines) != 10000 {
		t.Fatalf("got %d lines, want 10000", len(lines))
	}
	if lines[0] != "     1\tx" {
		t.Errorf("first line = %q", lines[0])
	}
	if lines[9] != "    10\tx" {
		t.Errorf("tenth line = %q", lines[9])
	}
	if lines[9999] != " 10000\tx" {
		t.Errorf("last line = %q", lines[9999])
	}
}

func TestE2E_NumberWidthExceedsSix(t *testing.T) {
	// 1,000,000 lines: the final number is 7 digits and must not be truncated.
	if testing.Short() {
		t.Skip("skipping large numbering test in short mode")
	}
	content := strings.Repeat("\n", 1000000)
	f := writeFile(t, "a.txt", content)
	stdout, _, _ := runCat(t, "-n", f)
	if !strings.HasSuffix(stdout, "1000000\t\n") {
		t.Errorf("last line not numbered as 1000000: %q", stdout[len(stdout)-20:])
	}
}

// ---------------------------------------------------------------------------
// -b / --number-nonblank
// ---------------------------------------------------------------------------

func TestE2E_NumberNonBlank(t *testing.T) {
	for _, flag := range []string{"-b", "--number-nonblank"} {
		t.Run(flag, func(t *testing.T) {
			f := writeFile(t, "a.txt", "a\n\n\nb\n\nc\n")
			stdout, _, code := runCat(t, flag, f)
			want := "     1\ta\n\n\n     2\tb\n\n     3\tc\n"
			if code != 0 || stdout != want {
				t.Errorf("code=%d stdout=%q want=%q", code, stdout, want)
			}
		})
	}
}

func TestE2E_NumberNonBlankOverridesNumber(t *testing.T) {
	for _, args := range [][]string{{"-bn"}, {"-nb"}, {"-b", "-n"}, {"-n", "-b"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			f := writeFile(t, "a.txt", "a\n\nb\n")
			stdout, _, _ := runCat(t, append(args, f)...)
			want := "     1\ta\n\n     2\tb\n"
			if stdout != want {
				t.Errorf("stdout=%q want=%q", stdout, want)
			}
		})
	}
}

func TestE2E_NumberNonBlankAcrossFiles(t *testing.T) {
	a := writeFile(t, "a.txt", "a\n\n")
	b := writeFile(t, "b.txt", "\nb\n")
	stdout, _, _ := runCat(t, "-b", a, b)
	if want := "     1\ta\n\n\n     2\tb\n"; stdout != want {
		t.Errorf("stdout=%q want=%q", stdout, want)
	}
}

func TestE2E_NumberNonBlankWhitespaceLineIsNonBlank(t *testing.T) {
	f := writeFile(t, "a.txt", "  \n\t\n\n")
	stdout, _, _ := runCat(t, "-b", f)
	if want := "     1\t  \n     2\t\t\n\n"; stdout != want {
		t.Errorf("stdout=%q want=%q", stdout, want)
	}
}

func TestE2E_NumberNonBlankNoTrailingNewline(t *testing.T) {
	f := writeFile(t, "a.txt", "a\n\nb")
	stdout, _, _ := runCat(t, "-b", f)
	if want := "     1\ta\n\n     2\tb"; stdout != want {
		t.Errorf("stdout=%q want=%q", stdout, want)
	}
}

// ---------------------------------------------------------------------------
// -s / --squeeze-blank
// ---------------------------------------------------------------------------

func TestE2E_SqueezeBlank(t *testing.T) {
	for _, flag := range []string{"-s", "--squeeze-blank"} {
		t.Run(flag, func(t *testing.T) {
			f := writeFile(t, "a.txt", "a\n\n\n\nb\n\n\nc\n")
			stdout, _, code := runCat(t, flag, f)
			want := "a\n\nb\n\nc\n"
			if code != 0 || stdout != want {
				t.Errorf("code=%d stdout=%q want=%q", code, stdout, want)
			}
		})
	}
}

func TestE2E_SqueezeBlankLeadingAndTrailing(t *testing.T) {
	f := writeFile(t, "a.txt", "\n\n\na\n\n\n")
	stdout, _, _ := runCat(t, "-s", f)
	if want := "\na\n\n"; stdout != want {
		t.Errorf("stdout=%q want=%q", stdout, want)
	}
}

func TestE2E_SqueezeBlankOnlyBlankLines(t *testing.T) {
	f := writeFile(t, "a.txt", "\n\n\n\n")
	stdout, _, _ := runCat(t, "-s", f)
	if want := "\n"; stdout != want {
		t.Errorf("stdout=%q want=%q", stdout, want)
	}
}

func TestE2E_SqueezeBlankKeepsSingleBlank(t *testing.T) {
	f := writeFile(t, "a.txt", "a\n\nb\n")
	stdout, _, _ := runCat(t, "-s", f)
	if stdout != "a\n\nb\n" {
		t.Errorf("stdout=%q", stdout)
	}
}

func TestE2E_SqueezeBlankWhitespaceLinesAreNotBlank(t *testing.T) {
	f := writeFile(t, "a.txt", "a\n \n \nb\n")
	stdout, _, _ := runCat(t, "-s", f)
	if stdout != "a\n \n \nb\n" {
		t.Errorf("stdout=%q", stdout)
	}
}

func TestE2E_SqueezeBlankWithNumber(t *testing.T) {
	f := writeFile(t, "a.txt", "a\n\n\n\nb\n")
	stdout, _, _ := runCat(t, "-sn", f)
	if want := "     1\ta\n     2\t\n     3\tb\n"; stdout != want {
		t.Errorf("stdout=%q want=%q", stdout, want)
	}
}

func TestE2E_SqueezeBlankWithNumberNonBlank(t *testing.T) {
	f := writeFile(t, "a.txt", "a\n\n\n\nb\n")
	stdout, _, _ := runCat(t, "-sb", f)
	if want := "     1\ta\n\n     2\tb\n"; stdout != want {
		t.Errorf("stdout=%q want=%q", stdout, want)
	}
}

func TestE2E_SqueezeBlankStdin(t *testing.T) {
	stdout, _, _ := runCatStdin(t, "a\n\n\n\nb\n", "-s")
	if stdout != "a\n\nb\n" {
		t.Errorf("stdout=%q", stdout)
	}
}

// ---------------------------------------------------------------------------
// -E / --show-ends
// ---------------------------------------------------------------------------

func TestE2E_ShowEnds(t *testing.T) {
	for _, flag := range []string{"-E", "--show-ends"} {
		t.Run(flag, func(t *testing.T) {
			f := writeFile(t, "a.txt", "a\n\nb\n")
			stdout, _, code := runCat(t, flag, f)
			if want := "a$\n$\nb$\n"; code != 0 || stdout != want {
				t.Errorf("code=%d stdout=%q want=%q", code, stdout, want)
			}
		})
	}
}

func TestE2E_ShowEndsNoTrailingNewline(t *testing.T) {
	f := writeFile(t, "a.txt", "a\nb")
	stdout, _, _ := runCat(t, "-E", f)
	if want := "a$\nb"; stdout != want {
		t.Errorf("stdout=%q want=%q", stdout, want)
	}
}

func TestE2E_ShowEndsCRLF(t *testing.T) {
	f := writeFile(t, "a.txt", "a\r\nb\r\n")
	stdout, _, _ := runCat(t, "-E", f)
	if want := "a\r$\nb\r$\n"; stdout != want {
		t.Errorf("stdout=%q want=%q", stdout, want)
	}
}

func TestE2E_ShowEndsWithNumber(t *testing.T) {
	f := writeFile(t, "a.txt", "a\n\n")
	stdout, _, _ := runCat(t, "-nE", f)
	if want := "     1\ta$\n     2\t$\n"; stdout != want {
		t.Errorf("stdout=%q want=%q", stdout, want)
	}
}

// ---------------------------------------------------------------------------
// -T / --show-tabs
// ---------------------------------------------------------------------------

func TestE2E_ShowTabs(t *testing.T) {
	for _, flag := range []string{"-T", "--show-tabs"} {
		t.Run(flag, func(t *testing.T) {
			f := writeFile(t, "a.txt", "a\tb\t\tc\n\tx\n")
			stdout, _, code := runCat(t, flag, f)
			if want := "a^Ib^I^Ic\n^Ix\n"; code != 0 || stdout != want {
				t.Errorf("code=%d stdout=%q want=%q", code, stdout, want)
			}
		})
	}
}

func TestE2E_ShowTabsLeavesNewlines(t *testing.T) {
	f := writeFile(t, "a.txt", "a\nb\n")
	stdout, _, _ := runCat(t, "-T", f)
	if stdout != "a\nb\n" {
		t.Errorf("stdout=%q", stdout)
	}
}

func TestE2E_NoShowTabsKeepsTabs(t *testing.T) {
	f := writeFile(t, "a.txt", "a\tb\n")
	stdout, _, _ := runCat(t, f)
	if stdout != "a\tb\n" {
		t.Errorf("stdout=%q", stdout)
	}
}

// ---------------------------------------------------------------------------
// -v / --show-nonprinting
// ---------------------------------------------------------------------------

func TestE2E_ShowNonPrintingControlChars(t *testing.T) {
	for _, flag := range []string{"-v", "--show-nonprinting"} {
		t.Run(flag, func(t *testing.T) {
			f := writeFile(t, "a.txt", "\x00\x01\x02\x07\x08\x0b\x0c\r\x1b\x1f\n")
			stdout, _, code := runCat(t, flag, f)
			want := "^@^A^B^G^H^K^L^M^[^_\n"
			if code != 0 || stdout != want {
				t.Errorf("code=%d stdout=%q want=%q", code, stdout, want)
			}
		})
	}
}

func TestE2E_ShowNonPrintingDEL(t *testing.T) {
	f := writeFile(t, "a.txt", "a\x7fb\n")
	stdout, _, _ := runCat(t, "-v", f)
	if want := "a^?b\n"; stdout != want {
		t.Errorf("stdout=%q want=%q", stdout, want)
	}
}

func TestE2E_ShowNonPrintingKeepsTabAndNewline(t *testing.T) {
	f := writeFile(t, "a.txt", "a\tb\nc\n")
	stdout, _, _ := runCat(t, "-v", f)
	if stdout != "a\tb\nc\n" {
		t.Errorf("stdout=%q", stdout)
	}
}

func TestE2E_ShowNonPrintingHighBit(t *testing.T) {
	// 0x80 -> M-^@, 0x9f -> M-^_, 0xa0 -> "M- ", 0xe9 -> M-i, 0xff -> M-^?
	f := writeFile(t, "a.txt", "\x80\x9f\xa0\xe9\xff\n")
	stdout, _, _ := runCat(t, "-v", f)
	if want := "M-^@M-^_M- M-iM-^?\n"; stdout != want {
		t.Errorf("stdout=%q want=%q", stdout, want)
	}
}

func TestE2E_ShowNonPrintingUTF8Bytes(t *testing.T) {
	// é = 0xC3 0xA9 -> M-CM-)
	f := writeFile(t, "a.txt", "é\n")
	stdout, _, _ := runCat(t, "-v", f)
	if want := "M-CM-)\n"; stdout != want {
		t.Errorf("stdout=%q want=%q", stdout, want)
	}
}

func TestE2E_ShowNonPrintingPrintableASCIIUntouched(t *testing.T) {
	var sb strings.Builder
	for c := 0x20; c < 0x7f; c++ {
		sb.WriteByte(byte(c))
	}
	sb.WriteByte('\n')
	f := writeFile(t, "a.txt", sb.String())
	stdout, _, _ := runCat(t, "-v", f)
	if stdout != sb.String() {
		t.Errorf("printable ASCII altered: %q", stdout)
	}
}

func TestE2E_ShowNonPrintingAllBytes(t *testing.T) {
	var in bytes.Buffer
	var want bytes.Buffer
	for i := 0; i < 256; i++ {
		b := byte(i)
		in.WriteByte(b)
		switch {
		case b == '\n' || b == '\t':
			want.WriteByte(b)
		case b < 32:
			want.WriteString("^" + string(rune(b+64)))
		case b < 127:
			want.WriteByte(b)
		case b == 127:
			want.WriteString("^?")
		case b < 160:
			want.WriteString("M-^" + string(rune(b-128+64)))
		case b < 255:
			want.WriteString("M-" + string(rune(b-128)))
		default:
			want.WriteString("M-^?")
		}
	}
	f := writeFile(t, "all.bin", in.String())
	stdout, _, _ := runCat(t, "-v", f)
	if stdout != want.String() {
		t.Errorf("-v byte table mismatch:\n got: %q\nwant: %q", stdout, want.String())
	}
}

// ---------------------------------------------------------------------------
// Composite flags: -A, -e, -t
// ---------------------------------------------------------------------------

func TestE2E_ShowAll(t *testing.T) {
	for _, flag := range []string{"-A", "--show-all"} {
		t.Run(flag, func(t *testing.T) {
			f := writeFile(t, "a.txt", "a\tb\x01\xff\r\n\n")
			stdout, _, code := runCat(t, flag, f)
			want := "a^Ib^AM-^?^M$\n$\n"
			if code != 0 || stdout != want {
				t.Errorf("code=%d stdout=%q want=%q", code, stdout, want)
			}
		})
	}
}

func TestE2E_ShowAllEquivalentToVET(t *testing.T) {
	f := writeFile(t, "a.txt", "x\ty\x00\x7f\x80\n\nz")
	a, _, _ := runCat(t, "-A", f)
	b, _, _ := runCat(t, "-vET", f)
	if a != b {
		t.Errorf("-A (%q) != -vET (%q)", a, b)
	}
}

func TestE2E_FlagE_EquivalentToVE(t *testing.T) {
	f := writeFile(t, "a.txt", "a\tb\x01\r\n")
	stdout, _, _ := runCat(t, "-e", f)
	if want := "a\tb^A^M$\n"; stdout != want {
		t.Errorf("stdout=%q want=%q", stdout, want)
	}
	ve, _, _ := runCat(t, "-vE", f)
	if stdout != ve {
		t.Errorf("-e (%q) != -vE (%q)", stdout, ve)
	}
}

func TestE2E_FlagT_EquivalentToVT(t *testing.T) {
	f := writeFile(t, "a.txt", "a\tb\x01\n")
	stdout, _, _ := runCat(t, "-t", f)
	if want := "a^Ib^A\n"; stdout != want {
		t.Errorf("stdout=%q want=%q", stdout, want)
	}
	vt, _, _ := runCat(t, "-vT", f)
	if stdout != vt {
		t.Errorf("-t (%q) != -vT (%q)", stdout, vt)
	}
}

func TestE2E_FlagU_Ignored(t *testing.T) {
	f := writeFile(t, "a.txt", "a\nb\n")
	stdout, _, code := runCat(t, "-u", f)
	if code != 0 || stdout != "a\nb\n" {
		t.Errorf("code=%d stdout=%q", code, stdout)
	}
}

// ---------------------------------------------------------------------------
// Option parsing
// ---------------------------------------------------------------------------

func TestE2E_CombinedShortFlags(t *testing.T) {
	f := writeFile(t, "a.txt", "a\tb\n\n\n\nc\n")
	stdout, _, code := runCat(t, "-nsET", f)
	want := "     1\ta^Ib$\n     2\t$\n     3\tc$\n"
	if code != 0 || stdout != want {
		t.Errorf("code=%d stdout=%q want=%q", code, stdout, want)
	}
}

func TestE2E_SeparateFlagsEqualCombined(t *testing.T) {
	f := writeFile(t, "a.txt", "a\tb\n\n\n\nc\x01\n")
	a, _, _ := runCat(t, "-n", "-s", "-E", "-T", "-v", f)
	b, _, _ := runCat(t, "-nsETv", f)
	if a != b {
		t.Errorf("separate (%q) != combined (%q)", a, b)
	}
}

func TestE2E_LongFlagsCombined(t *testing.T) {
	f := writeFile(t, "a.txt", "a\tb\n\n\n\nc\n")
	stdout, _, code := runCat(t, "--number", "--squeeze-blank", "--show-ends", "--show-tabs", f)
	want := "     1\ta^Ib$\n     2\t$\n     3\tc$\n"
	if code != 0 || stdout != want {
		t.Errorf("code=%d stdout=%q want=%q", code, stdout, want)
	}
}

func TestE2E_FlagsAfterFilenames(t *testing.T) {
	f := writeFile(t, "a.txt", "a\n")
	stdout, _, code := runCat(t, f, "-n")
	if code != 0 || stdout != "     1\ta\n" {
		t.Errorf("code=%d stdout=%q", code, stdout)
	}
}

func TestE2E_DoubleDashEndsOptions(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "-n")
	if err := os.WriteFile(name, []byte("dash file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(binaryPath, "--", "-n")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("cat -- -n failed: %v", err)
	}
	if string(out) != "dash file\n" {
		t.Errorf("stdout=%q", out)
	}
}

func TestE2E_UnknownShortOption(t *testing.T) {
	stdout, stderr, code := runCat(t, "-Z")
	if code == 0 {
		t.Error("expected non-zero exit for unknown option")
	}
	if stdout != "" {
		t.Errorf("stdout=%q, want empty", stdout)
	}
	if stderr == "" && !strings.Contains(stdout, "Usage") {
		t.Error("expected an error message for unknown option")
	}
}

func TestE2E_UnknownLongOption(t *testing.T) {
	_, stderr, code := runCat(t, "--bogus")
	if code == 0 {
		t.Error("expected non-zero exit for unknown option")
	}
	if !strings.Contains(stderr, "bogus") {
		t.Errorf("stderr should mention the bad option, got %q", stderr)
	}
}

func TestE2E_HelpDoesNotAcceptShortH(t *testing.T) {
	// GNU cat has no -h option.
	_, _, code := runCat(t, "-h")
	if code == 0 {
		t.Error("-h should be rejected like GNU cat")
	}
}

func TestE2E_VIsShowNonPrintingNotVersion(t *testing.T) {
	f := writeFile(t, "a.txt", "a\x01\n")
	stdout, _, code := runCat(t, "-v", f)
	if code != 0 || stdout != "a^A\n" {
		t.Errorf("-v must be --show-nonprinting; code=%d stdout=%q", code, stdout)
	}
}

// ---------------------------------------------------------------------------
// Error handling
// ---------------------------------------------------------------------------

func TestE2E_MissingFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist.txt")
	stdout, stderr, code := runCat(t, missing)
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if stdout != "" {
		t.Errorf("stdout=%q, want empty", stdout)
	}
	if !strings.Contains(stderr, "does-not-exist.txt") || !strings.HasPrefix(stderr, "cat: ") {
		t.Errorf("stderr=%q, want 'cat: <file>: ...'", stderr)
	}
}

func TestE2E_MissingFileContinuesWithOthers(t *testing.T) {
	a := writeFile(t, "a.txt", "a\n")
	b := writeFile(t, "b.txt", "b\n")
	missing := filepath.Join(t.TempDir(), "missing.txt")
	stdout, stderr, code := runCat(t, a, missing, b)
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if stdout != "a\nb\n" {
		t.Errorf("stdout=%q, want %q", stdout, "a\nb\n")
	}
	if !strings.Contains(stderr, "missing.txt") {
		t.Errorf("stderr=%q", stderr)
	}
}

func TestE2E_MultipleMissingFiles(t *testing.T) {
	dir := t.TempDir()
	_, stderr, code := runCat(t, filepath.Join(dir, "x1"), filepath.Join(dir, "x2"))
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "x1") || !strings.Contains(stderr, "x2") {
		t.Errorf("stderr should report both files, got %q", stderr)
	}
}

func TestE2E_Directory(t *testing.T) {
	dir := t.TempDir()
	stdout, stderr, code := runCat(t, dir)
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if stdout != "" {
		t.Errorf("stdout=%q, want empty", stdout)
	}
	if !strings.HasPrefix(stderr, "cat: ") {
		t.Errorf("stderr=%q", stderr)
	}
}

func TestE2E_DirectoryContinuesWithOthers(t *testing.T) {
	a := writeFile(t, "a.txt", "a\n")
	stdout, _, code := runCat(t, t.TempDir(), a)
	if code != 1 || stdout != "a\n" {
		t.Errorf("code=%d stdout=%q", code, stdout)
	}
}

func TestE2E_NumberingUnaffectedByMissingFile(t *testing.T) {
	a := writeFile(t, "a.txt", "a\n")
	b := writeFile(t, "b.txt", "b\n")
	missing := filepath.Join(t.TempDir(), "missing")
	stdout, _, code := runCat(t, "-n", a, missing, b)
	if code != 1 || stdout != "     1\ta\n     2\tb\n" {
		t.Errorf("code=%d stdout=%q", code, stdout)
	}
}

func TestE2E_EmptyFilenameIsError(t *testing.T) {
	_, stderr, code := runCat(t, "")
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.HasPrefix(stderr, "cat: ") {
		t.Errorf("stderr=%q", stderr)
	}
}

func TestE2E_SuccessExitCodeAndEmptyStderr(t *testing.T) {
	f := writeFile(t, "a.txt", "ok\n")
	_, stderr, code := runCat(t, f)
	if code != 0 || stderr != "" {
		t.Errorf("code=%d stderr=%q", code, stderr)
	}
}

// ---------------------------------------------------------------------------
// Closed output pipe (e.g. `cat big | head -n 1`)
// ---------------------------------------------------------------------------

func TestE2E_ClosedPipeTerminates(t *testing.T) {
	content := strings.Repeat("a line of text\n", 200000) // ~3 MB
	f := writeFile(t, "big.txt", content)

	cmd := exec.Command(binaryPath, f)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 16)
	if _, err := io.ReadFull(stdout, buf); err != nil {
		t.Fatalf("failed to read from cat: %v", err)
	}
	_ = stdout.Close()

	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("cat did not terminate after its output pipe was closed")
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkCat_Throughput(b *testing.B) {
	content := strings.Repeat("The quick brown fox jumps over the lazy dog\n", 250000) // ~11 MB
	path := filepath.Join(b.TempDir(), "bench.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(content)))
	for b.Loop() {
		cmd := exec.Command(binaryPath, path)
		cmd.Stdout = io.Discard
		if err := cmd.Run(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCat_NumberAndShowAll(b *testing.B) {
	content := strings.Repeat("The\tquick\x01 brown fox\n\n", 100000)
	path := filepath.Join(b.TempDir(), "bench.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(content)))
	for b.Loop() {
		cmd := exec.Command(binaryPath, "-nA", path)
		cmd.Stdout = io.Discard
		if err := cmd.Run(); err != nil {
			b.Fatal(err)
		}
	}
}
