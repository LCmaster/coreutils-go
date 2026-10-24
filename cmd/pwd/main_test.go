package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/LCmaster/coreutils-go/pkg/fsutils"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "pwd-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binName := "pwd"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	bin := filepath.Join(tmpDir, binName)

	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build pwd binary: %v\n%s\n", err, out)
		os.Exit(1)
	}
	binaryPath = bin

	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func runPwd(t *testing.T, workDir string, env []string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	if env != nil {
		cmd.Env = env
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run pwd: %v", err)
		}
	}
	return stdout.String(), stderr.String(), code
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
		_, _ = io.Copy(&buf, r)
		outChan <- buf.String()
	}()

	ferr := f()
	_ = w.Close()
	os.Stdout = oldStdout
	out := <-outChan
	_ = r.Close()

	if ferr != nil {
		return out, ferr
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// Unit tests: Help & Version
// ---------------------------------------------------------------------------

func TestUnit_PrintHelp(t *testing.T) {
	help := getHelpStr()
	for _, expected := range []string{
		"Usage:",
		"pwd [OPTION]...",
		"Description:",
		"Print the full filename of the current working directory.",
		"-L, --logical",
		"-P, --physical",
		"-h, --help",
		"-v, --version",
	} {
		if !strings.Contains(help, expected) {
			t.Errorf("getHelpStr() missing snippet %q", expected)
		}
	}
	if !strings.HasSuffix(help, "\n") {
		t.Error("getHelpStr() should end with a newline")
	}
}

func TestUnit_PrintVersion(t *testing.T) {
	ver := getVersionStr()
	if ver != "pwd (coreutils-go) 0.1.0" {
		t.Errorf("getVersionStr() = %q, want 'pwd (coreutils-go) 0.1.0'", ver)
	}
}

// ---------------------------------------------------------------------------
// Unit tests: Functional Combinators & Predicates
// ---------------------------------------------------------------------------

func TestUnit_Functional_SplitPathSegments(t *testing.T) {
	segments := fsutils.SplitSegments("/a/b//c/d/")
	expected := []string{"a", "b", "c", "d"}
	if len(segments) != len(expected) {
		t.Fatalf("splitPathSegments() got %v, want %v", segments, expected)
	}
	for i, seg := range segments {
		if seg != expected[i] {
			t.Errorf("segment[%d] = %q, want %q", i, seg, expected[i])
		}
	}
}

func TestUnit_HasDotComponents(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/a/b/c", false},
		{"/a/./b", true},
		{"/a/../b", true},
		{".", true},
		{"..", true},
		{"", false},
	}
	for _, tt := range tests {
		got := hasDotComponents(tt.path)
		if got != tt.want {
			t.Errorf("hasDotComponents(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestUnit_Functional_AllPredicates(t *testing.T) {
	alwaysTrue := func(string) bool { return true }
	alwaysFalse := func(string) bool { return false }

	if !allPredicates(alwaysTrue, alwaysTrue)("test") {
		t.Error("allPredicates(true, true) should be true")
	}
	if allPredicates(alwaysTrue, alwaysFalse)("test") {
		t.Error("allPredicates(true, false) should be false")
	}
	if allPredicates(alwaysFalse, alwaysTrue)("test") {
		t.Error("allPredicates(false, true) should be false")
	}
	if !allPredicates()("test") {
		t.Error("empty allPredicates() should be true")
	}
}

func TestUnit_ResolveLogicalDir_FallsBackOnInvalidPWD(t *testing.T) {
	physical, err := resolvePhysicalDir()
	if err != nil {
		t.Fatalf("resolvePhysicalDir() error: %v", err)
	}

	for _, invalid := range []string{"", "relative/path", "/definitely/not/a/real/dir"} {
		t.Setenv("PWD", invalid)
		got, err := resolveLogicalDir()
		if err != nil {
			t.Fatalf("PWD=%q: resolveLogicalDir() error: %v", invalid, err)
		}
		if got != physical {
			t.Errorf("PWD=%q: resolveLogicalDir() = %q, want physical %q", invalid, got, physical)
		}
	}
}

func TestUnit_ResolveLogicalDir_UsesValidPWD(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error: %v", err)
	}
	t.Setenv("PWD", cwd)
	got, err := resolveLogicalDir()
	if err != nil || got != cwd {
		t.Errorf("resolveLogicalDir() = %q, %v; want %q", got, err, cwd)
	}
}

// ---------------------------------------------------------------------------
// Unit tests: Directory Resolvers
// ---------------------------------------------------------------------------

func TestUnit_ResolvePhysicalDir(t *testing.T) {
	dir, err := resolvePhysicalDir()
	if err != nil {
		t.Fatalf("resolvePhysicalDir() error: %v", err)
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("resolvePhysicalDir() returned non-absolute path: %q", dir)
	}
}

func TestUnit_ResolveLogicalDir(t *testing.T) {
	dir, err := resolveLogicalDir()
	if err != nil {
		t.Fatalf("resolveLogicalDir() error: %v", err)
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("resolveLogicalDir() returned non-absolute path: %q", dir)
	}
}

func TestUnit_ResolveWorkingDir(t *testing.T) {
	logical, err := resolveWorkingDir(LogicalMode)
	if err != nil {
		t.Fatalf("resolveWorkingDir(LogicalMode) error: %v", err)
	}
	physical, err := resolveWorkingDir(PhysicalMode)
	if err != nil {
		t.Fatalf("resolveWorkingDir(PhysicalMode) error: %v", err)
	}
	if logical == "" || physical == "" {
		t.Error("working directory should not be empty")
	}
}

// ---------------------------------------------------------------------------
// Unit tests: Option Parsing & Execution Runner
// ---------------------------------------------------------------------------

func TestUnit_ParseArgs(t *testing.T) {
	// Default mode is Logical
	opts, err := parseArgs([]string{})
	if err != nil || opts.Mode != LogicalMode || opts.Action != ActionRun {
		t.Errorf("parseArgs([]) = %+v, %v; want LogicalMode, ActionRun", opts, err)
	}

	// -L flag
	opts, err = parseArgs([]string{"-L"})
	if err != nil || opts.Mode != LogicalMode {
		t.Errorf("parseArgs(-L) = %+v, %v; want LogicalMode", opts, err)
	}

	// -P flag
	opts, err = parseArgs([]string{"-P"})
	if err != nil || opts.Mode != PhysicalMode {
		t.Errorf("parseArgs(-P) = %+v, %v; want PhysicalMode", opts, err)
	}

	// Long flags
	opts, err = parseArgs([]string{"--physical"})
	if err != nil || opts.Mode != PhysicalMode {
		t.Errorf("parseArgs(--physical) = %+v, %v; want PhysicalMode", opts, err)
	}

	opts, err = parseArgs([]string{"--logical"})
	if err != nil || opts.Mode != LogicalMode {
		t.Errorf("parseArgs(--logical) = %+v, %v; want LogicalMode", opts, err)
	}

	// Precedence: -L then -P -> Physical
	opts, err = parseArgs([]string{"-L", "-P"})
	if err != nil || opts.Mode != PhysicalMode {
		t.Errorf("parseArgs(-L, -P) = %+v, %v; want PhysicalMode", opts, err)
	}

	// Precedence: -P then -L -> Logical
	opts, err = parseArgs([]string{"-P", "-L"})
	if err != nil || opts.Mode != LogicalMode {
		t.Errorf("parseArgs(-P, -L) = %+v, %v; want LogicalMode", opts, err)
	}

	// Bundled: -LP -> Physical
	opts, err = parseArgs([]string{"-LP"})
	if err != nil || opts.Mode != PhysicalMode {
		t.Errorf("parseArgs(-LP) = %+v, %v; want PhysicalMode", opts, err)
	}

	// Bundled: -PL -> Logical
	opts, err = parseArgs([]string{"-PL"})
	if err != nil || opts.Mode != LogicalMode {
		t.Errorf("parseArgs(-PL) = %+v, %v; want LogicalMode", opts, err)
	}

	// Help and version
	opts, err = parseArgs([]string{"-h"})
	if err != nil || opts.Action != ActionHelp {
		t.Errorf("parseArgs(-h) = %+v, %v; want ActionHelp", opts, err)
	}

	opts, err = parseArgs([]string{"--version"})
	if err != nil || opts.Action != ActionVersion {
		t.Errorf("parseArgs(--version) = %+v, %v; want ActionVersion", opts, err)
	}

	// Double-dash terminates options
	opts, err = parseArgs([]string{"--", "-P"})
	if err != nil || opts.Mode != LogicalMode {
		t.Errorf("parseArgs(--, -P) = %+v, %v; want LogicalMode (default)", opts, err)
	}

	// Invalid short flag
	_, err = parseArgs([]string{"-x"})
	if err == nil || !strings.Contains(err.Error(), "invalid option -- 'x'") {
		t.Errorf("expected 'invalid option -- 'x'', got %v", err)
	}

	// Unrecognized long option
	_, err = parseArgs([]string{"--invalid-flag"})
	if err == nil || !strings.Contains(err.Error(), "unrecognized option '--invalid-flag'") {
		t.Errorf("expected 'unrecognized option '--invalid-flag'', got %v", err)
	}
}

func TestUnit_Run(t *testing.T) {
	// Test Help
	var outHelp, errHelp bytes.Buffer
	code := run([]string{"-h"}, nil, &outHelp, &errHelp)
	if code != 0 || !strings.Contains(outHelp.String(), "Usage:") || errHelp.Len() > 0 {
		t.Errorf("run(-h) code=%d, stdout=%q, stderr=%q", code, outHelp.String(), errHelp.String())
	}

	// Test Version
	var outVer, errVer bytes.Buffer
	code = run([]string{"--version"}, nil, &outVer, &errVer)
	if code != 0 || outVer.String() != "pwd (coreutils-go) 0.1.0\n" || errVer.Len() > 0 {
		t.Errorf("run(--version) code=%d, stdout=%q, stderr=%q", code, outVer.String(), errVer.String())
	}

	// Test Invalid Option
	var outInv, errInv bytes.Buffer
	code = run([]string{"--not-a-flag"}, nil, &outInv, &errInv)
	if code != 1 || !strings.Contains(errInv.String(), "unrecognized option '--not-a-flag'") {
		t.Errorf("run(--not-a-flag) code=%d, stderr=%q", code, errInv.String())
	}

	// Test Success Run
	var outRun, errRun bytes.Buffer
	code = run([]string{}, nil, &outRun, &errRun)
	if code != 0 || outRun.Len() == 0 || errRun.Len() > 0 {
		t.Errorf("run() code=%d, stdout=%q, stderr=%q", code, outRun.String(), errRun.String())
	}
}

func TestUnit_IsBrokenPipe(t *testing.T) {
	if isBrokenPipe(nil) {
		t.Error("nil should not be broken pipe")
	}
	if !isBrokenPipe(fmt.Errorf("write: broken pipe")) {
		t.Error("expected broken pipe to be detected")
	}
}

// ---------------------------------------------------------------------------
// End-to-End (CLI) integration tests
// ---------------------------------------------------------------------------

func TestE2E_Basic(t *testing.T) {
	tmp := t.TempDir()
	stdout, stderr, code := runPwd(t, tmp, nil)
	if code != 0 || stderr != "" {
		t.Fatalf("code=%d, stderr=%q", code, stderr)
	}

	expected, err := filepath.EvalSymlinks(tmp)
	if err != nil {
		expected = tmp
	}
	gotClean := filepath.Clean(strings.TrimSpace(stdout))
	expectedClean := filepath.Clean(expected)
	if !strings.EqualFold(gotClean, expectedClean) {
		t.Errorf("stdout = %q, want %q", gotClean, expectedClean)
	}
}

func TestE2E_Help(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		t.Run(flag, func(t *testing.T) {
			stdout, stderr, code := runPwd(t, "", nil, flag)
			if code != 0 || stderr != "" {
				t.Fatalf("code=%d, stderr=%q", code, stderr)
			}
			if stdout != getHelpStr() {
				t.Errorf("%s output mismatch:\n got: %q\nwant: %q", flag, stdout, getHelpStr())
			}
		})
	}
}

func TestE2E_Version(t *testing.T) {
	for _, flag := range []string{"--version", "-v"} {
		t.Run(flag, func(t *testing.T) {
			stdout, stderr, code := runPwd(t, "", nil, flag)
			if code != 0 || stderr != "" {
				t.Fatalf("code=%d, stderr=%q", code, stderr)
			}
			want := "pwd (coreutils-go) 0.1.0\n"
			if stdout != want {
				t.Errorf("%s output = %q, want %q", flag, stdout, want)
			}
		})
	}
}

func TestE2E_PhysicalFlag(t *testing.T) {
	tmp := t.TempDir()
	for _, flag := range []string{"-P", "--physical"} {
		t.Run(flag, func(t *testing.T) {
			stdout, stderr, code := runPwd(t, tmp, nil, flag)
			if code != 0 || stderr != "" {
				t.Fatalf("code=%d, stderr=%q", code, stderr)
			}
			expected, err := filepath.EvalSymlinks(tmp)
			if err != nil {
				expected = tmp
			}
			if !strings.EqualFold(filepath.Clean(strings.TrimSpace(stdout)), filepath.Clean(expected)) {
				t.Errorf("stdout=%q, want=%q", strings.TrimSpace(stdout), expected)
			}
		})
	}
}

func TestE2E_LogicalFlag(t *testing.T) {
	tmp := t.TempDir()
	for _, flag := range []string{"-L", "--logical"} {
		t.Run(flag, func(t *testing.T) {
			stdout, stderr, code := runPwd(t, tmp, nil, flag)
			if code != 0 || stderr != "" {
				t.Fatalf("code=%d, stderr=%q", code, stderr)
			}
			if strings.TrimSpace(stdout) == "" {
				t.Error("expected non-empty output")
			}
		})
	}
}

func TestE2E_FlagPrecedence(t *testing.T) {
	tmp := t.TempDir()

	// -L followed by -P should act as -P
	stdoutP, stderr, code := runPwd(t, tmp, nil, "-L", "-P")
	if code != 0 || stderr != "" {
		t.Fatalf("-L -P failed: code=%d, stderr=%q", code, stderr)
	}

	stdoutSoloP, _, _ := runPwd(t, tmp, nil, "-P")
	if stdoutP != stdoutSoloP {
		t.Errorf("'-L -P' (%q) != '-P' (%q)", stdoutP, stdoutSoloP)
	}

	// -P followed by -L should act as -L
	stdoutL, stderr, code := runPwd(t, tmp, nil, "-P", "-L")
	if code != 0 || stderr != "" {
		t.Fatalf("-P -L failed: code=%d, stderr=%q", code, stderr)
	}

	stdoutSoloL, _, _ := runPwd(t, tmp, nil, "-L")
	if stdoutL != stdoutSoloL {
		t.Errorf("'-P -L' (%q) != '-L' (%q)", stdoutL, stdoutSoloL)
	}

	// Combined short flags
	stdoutLP, _, _ := runPwd(t, tmp, nil, "-LP")
	if stdoutLP != stdoutSoloP {
		t.Errorf("'-LP' (%q) != '-P' (%q)", stdoutLP, stdoutSoloP)
	}

	stdoutPL, _, _ := runPwd(t, tmp, nil, "-PL")
	if stdoutPL != stdoutSoloL {
		t.Errorf("'-PL' (%q) != '-L' (%q)", stdoutPL, stdoutSoloL)
	}
}

func TestE2E_SymlinkResolution(t *testing.T) {
	realDir := t.TempDir()
	symlinkDir := filepath.Join(t.TempDir(), "link_to_real")

	err := os.Symlink(realDir, symlinkDir)
	if err != nil {
		t.Skipf("skipping symlink test: symlinks not permitted or supported on this system: %v", err)
	}

	// Physical mode must resolve to the real directory target
	stdoutP, stderr, code := runPwd(t, symlinkDir, nil, "-P")
	if code != 0 || stderr != "" {
		t.Fatalf("pwd -P failed in symlink: code=%d, stderr=%q", code, stderr)
	}

	expectedPhysical, err := filepath.EvalSymlinks(realDir)
	if err != nil {
		expectedPhysical = realDir
	}
	if !strings.EqualFold(filepath.Clean(strings.TrimSpace(stdoutP)), filepath.Clean(expectedPhysical)) {
		t.Errorf("pwd -P = %q, want canonical %q", strings.TrimSpace(stdoutP), expectedPhysical)
	}
}

func TestE2E_ExtraArguments(t *testing.T) {
	// POSIX/GNU pwd ignores extra operands
	tmp := t.TempDir()
	stdout, stderr, code := runPwd(t, tmp, nil, "extra", "operands")
	if code != 0 || stderr != "" {
		t.Fatalf("code=%d, stderr=%q", code, stderr)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Error("expected non-empty output")
	}
}

func TestE2E_InvalidFlag(t *testing.T) {
	_, stderr, code := runPwd(t, "", nil, "--invalid-flag")
	if code == 0 {
		t.Error("expected non-zero exit code for invalid flag")
	}
	if stderr == "" {
		t.Error("expected error message on stderr for invalid flag")
	}
}
