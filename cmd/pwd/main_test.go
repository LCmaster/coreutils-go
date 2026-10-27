package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/LCmaster/coreutils-go/pkg/cli"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "pwd-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	binName := "pwd"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	bin := filepath.Join(tmpDir, binName)

	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build pwd binary: %v\n%s\n", err, out)
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

// ---------------------------------------------------------------------------
// Unit tests: Help & Version
// ---------------------------------------------------------------------------

func TestUnit_PrintHelp(t *testing.T) {
	help := newCommand().Help()
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
			t.Errorf("cmd.Help() missing snippet %q", expected)
		}
	}
	if !strings.HasSuffix(help, "\n") {
		t.Error("cmd.Help() should end with a newline")
	}
}

func TestUnit_PrintVersion(t *testing.T) {
	ver := newCommand().Version()
	if ver != "pwd (coreutils-go) 0.1.0" {
		t.Errorf("cmd.Version() = %q, want 'pwd (coreutils-go) 0.1.0'", ver)
	}
}

// ---------------------------------------------------------------------------
// Unit tests: Directory Resolvers
// ---------------------------------------------------------------------------

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
	logical, err := resolveWorkingDir(false)
	if err != nil {
		t.Fatalf("resolveWorkingDir(false) error: %v", err)
	}
	physical, err := resolveWorkingDir(true)
	if err != nil {
		t.Fatalf("resolveWorkingDir(true) error: %v", err)
	}
	if logical == "" || physical == "" {
		t.Error("working directory should not be empty")
	}
}

// ---------------------------------------------------------------------------
// Unit tests: Option Parsing & Execution Runner
// ---------------------------------------------------------------------------

func TestUnit_Options(t *testing.T) {
	// Default mode is Logical (physical = false)
	cmd := newCommand()
	err := cmd.Parse([]string{})
	if err != nil || cmd.Options.Bools["physical"] || cmd.Action() != cli.ActionRun {
		t.Errorf("default parse error=%v, physical=%v, action=%v", err, cmd.Options.Bools["physical"], cmd.Action())
	}

	// -L flag
	cmd = newCommand()
	err = cmd.Parse([]string{"-L"})
	if err != nil || cmd.Options.Bools["physical"] {
		t.Errorf("-L parse error=%v, physical=%v", err, cmd.Options.Bools["physical"])
	}

	// -P flag
	cmd = newCommand()
	err = cmd.Parse([]string{"-P"})
	if err != nil || !cmd.Options.Bools["physical"] {
		t.Errorf("-P parse error=%v, physical=%v", err, cmd.Options.Bools["physical"])
	}

	// Long flags
	cmd = newCommand()
	err = cmd.Parse([]string{"--physical"})
	if err != nil || !cmd.Options.Bools["physical"] {
		t.Errorf("--physical parse error=%v, physical=%v", err, cmd.Options.Bools["physical"])
	}

	cmd = newCommand()
	err = cmd.Parse([]string{"--logical"})
	if err != nil || cmd.Options.Bools["physical"] {
		t.Errorf("--logical parse error=%v, physical=%v", err, cmd.Options.Bools["physical"])
	}

	// Precedence: -L then -P -> Physical
	cmd = newCommand()
	err = cmd.Parse([]string{"-L", "-P"})
	if err != nil || !cmd.Options.Bools["physical"] {
		t.Errorf("-L -P parse error=%v, physical=%v", err, cmd.Options.Bools["physical"])
	}

	// Precedence: -P then -L -> Logical
	cmd = newCommand()
	err = cmd.Parse([]string{"-P", "-L"})
	if err != nil || cmd.Options.Bools["physical"] {
		t.Errorf("-P -L parse error=%v, physical=%v", err, cmd.Options.Bools["physical"])
	}

	// Bundled: -LP -> Physical
	cmd = newCommand()
	err = cmd.Parse([]string{"-LP"})
	if err != nil || !cmd.Options.Bools["physical"] {
		t.Errorf("-LP parse error=%v, physical=%v", err, cmd.Options.Bools["physical"])
	}

	// Bundled: -PL -> Logical
	cmd = newCommand()
	err = cmd.Parse([]string{"-PL"})
	if err != nil || cmd.Options.Bools["physical"] {
		t.Errorf("-PL parse error=%v, physical=%v", err, cmd.Options.Bools["physical"])
	}

	// Help and version
	cmd = newCommand()
	err = cmd.Parse([]string{"-h"})
	if err != nil || cmd.Action() != cli.ActionHelp {
		t.Errorf("parse(-h) = %v, want ActionHelp", cmd.Action())
	}

	cmd = newCommand()
	err = cmd.Parse([]string{"--version"})
	if err != nil || cmd.Action() != cli.ActionVersion {
		t.Errorf("parse(--version) = %v, want ActionVersion", cmd.Action())
	}

	// Double-dash terminates options
	cmd = newCommand()
	err = cmd.Parse([]string{"--", "-P"})
	if err != nil || cmd.Options.Bools["physical"] {
		t.Errorf("parse(--, -P) physical=%v, want false", cmd.Options.Bools["physical"])
	}

	// Invalid short flag
	cmd = newCommand()
	err = cmd.Parse([]string{"-x"})
	if err == nil || !strings.Contains(err.Error(), "invalid option -- 'x'") {
		t.Errorf("expected 'invalid option -- 'x'', got %v", err)
	}

	// Unrecognized long option
	cmd = newCommand()
	err = cmd.Parse([]string{"--invalid-flag"})
	if err == nil || !strings.Contains(err.Error(), "unrecognized option '--invalid-flag'") {
		t.Errorf("expected 'unrecognized option '--invalid-flag'', got %v", err)
	}
}

func TestUnit_Run(t *testing.T) {
	// Test Help
	var outHelp, errHelp bytes.Buffer
	err := newCommand().Execute([]string{"-h"}, nil, &outHelp, &errHelp)
	if err != nil || !strings.Contains(outHelp.String(), "Usage:") || errHelp.Len() > 0 {
		t.Errorf("Execute(-h) err=%v, stdout=%q, stderr=%q", err, outHelp.String(), errHelp.String())
	}

	// Test Version
	var outVer, errVer bytes.Buffer
	err = newCommand().Execute([]string{"--version"}, nil, &outVer, &errVer)
	if err != nil || outVer.String() != "pwd (coreutils-go) 0.1.0\n" || errVer.Len() > 0 {
		t.Errorf("Execute(--version) err=%v, stdout=%q, stderr=%q", err, outVer.String(), errVer.String())
	}

	// Test Invalid Option
	var outInv, errInv bytes.Buffer
	err = newCommand().Execute([]string{"--not-a-flag"}, nil, &outInv, &errInv)
	if err == nil || !strings.Contains(errInv.String(), "unrecognized option '--not-a-flag'") {
		t.Errorf("Execute(--not-a-flag) err=%v, stderr=%q", err, errInv.String())
	}

	// Test Success Run
	var outRun, errRun bytes.Buffer
	err = newCommand().Execute([]string{}, nil, &outRun, &errRun)
	if err != nil || outRun.Len() == 0 || errRun.Len() > 0 {
		t.Errorf("Execute() err=%v, stdout=%q, stderr=%q", err, outRun.String(), errRun.String())
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
			expectedHelp := newCommand().Help()
			if stdout != expectedHelp {
				t.Errorf("%s output mismatch:\n got: %q\nwant: %q", flag, stdout, expectedHelp)
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
