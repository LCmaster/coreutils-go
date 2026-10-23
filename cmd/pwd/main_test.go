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
// Unit tests
// ---------------------------------------------------------------------------

func TestUnit_PrintHelp(t *testing.T) {
	help := printHelp()
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
			t.Errorf("printHelp() missing snippet %q", expected)
		}
	}
	if !strings.HasSuffix(help, "\n") {
		t.Error("printHelp() should end with a newline")
	}
}

func TestUnit_PrintVersion(t *testing.T) {
	out, err := captureStdout(printVersion)
	if err != nil {
		t.Fatalf("printVersion error: %v", err)
	}
	if out != "pwd (coreutils-go) 0.1.0\n" {
		t.Errorf("printVersion() = %q, want 'pwd (coreutils-go) 0.1.0\\n'", out)
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

func TestUnit_GetPhysicalDir(t *testing.T) {
	dir, err := getPhysicalDir()
	if err != nil {
		t.Fatalf("getPhysicalDir() error: %v", err)
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("getPhysicalDir() returned non-absolute path: %q", dir)
	}
}

func TestUnit_GetLogicalDir(t *testing.T) {
	dir, err := getLogicalDir()
	if err != nil {
		t.Fatalf("getLogicalDir() error: %v", err)
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("getLogicalDir() returned non-absolute path: %q", dir)
	}
}

func TestUnit_GetWorkingDir(t *testing.T) {
	logical, err := getWorkingDir(LogicalMode)
	if err != nil {
		t.Fatalf("getWorkingDir(LogicalMode) error: %v", err)
	}
	physical, err := getWorkingDir(PhysicalMode)
	if err != nil {
		t.Fatalf("getWorkingDir(PhysicalMode) error: %v", err)
	}
	if logical == "" || physical == "" {
		t.Error("working directory should not be empty")
	}
}

func TestUnit_ParseArgs(t *testing.T) {
	// Default mode is Logical
	res, err := parseArgs([]string{})
	if err != nil || res.Mode != LogicalMode || res.Action != ActionRun {
		t.Errorf("parseArgs([]) = %+v, %v; want LogicalMode, ActionRun", res, err)
	}

	// -L flag
	res, err = parseArgs([]string{"-L"})
	if err != nil || res.Mode != LogicalMode {
		t.Errorf("parseArgs(-L) = %+v, %v; want LogicalMode", res, err)
	}

	// -P flag
	res, err = parseArgs([]string{"-P"})
	if err != nil || res.Mode != PhysicalMode {
		t.Errorf("parseArgs(-P) = %+v, %v; want PhysicalMode", res, err)
	}

	// Long flags
	res, err = parseArgs([]string{"--physical"})
	if err != nil || res.Mode != PhysicalMode {
		t.Errorf("parseArgs(--physical) = %+v, %v; want PhysicalMode", res, err)
	}

	res, err = parseArgs([]string{"--logical"})
	if err != nil || res.Mode != LogicalMode {
		t.Errorf("parseArgs(--logical) = %+v, %v; want LogicalMode", res, err)
	}

	// Precedence: -L then -P -> Physical
	res, err = parseArgs([]string{"-L", "-P"})
	if err != nil || res.Mode != PhysicalMode {
		t.Errorf("parseArgs(-L, -P) = %+v, %v; want PhysicalMode", res, err)
	}

	// Precedence: -P then -L -> Logical
	res, err = parseArgs([]string{"-P", "-L"})
	if err != nil || res.Mode != LogicalMode {
		t.Errorf("parseArgs(-P, -L) = %+v, %v; want LogicalMode", res, err)
	}

	// Bundled: -LP -> Physical
	res, err = parseArgs([]string{"-LP"})
	if err != nil || res.Mode != PhysicalMode {
		t.Errorf("parseArgs(-LP) = %+v, %v; want PhysicalMode", res, err)
	}

	// Bundled: -PL -> Logical
	res, err = parseArgs([]string{"-PL"})
	if err != nil || res.Mode != LogicalMode {
		t.Errorf("parseArgs(-PL) = %+v, %v; want LogicalMode", res, err)
	}

	// Help and version
	res, err = parseArgs([]string{"-h"})
	if err != nil || res.Action != ActionHelp {
		t.Errorf("parseArgs(-h) = %+v, %v; want ActionHelp", res, err)
	}

	res, err = parseArgs([]string{"--version"})
	if err != nil || res.Action != ActionVersion {
		t.Errorf("parseArgs(--version) = %+v, %v; want ActionVersion", res, err)
	}

	// Double-dash terminates options
	res, err = parseArgs([]string{"--", "-P"})
	if err != nil || res.Mode != LogicalMode {
		t.Errorf("parseArgs(--, -P) = %+v, %v; want LogicalMode (default)", res, err)
	}

	// Invalid flag
	_, err = parseArgs([]string{"-x"})
	if err == nil {
		t.Error("expected error for unknown flag -x")
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
			if stdout != printHelp() {
				t.Errorf("%s output mismatch:\n got: %q\nwant: %q", flag, stdout, printHelp())
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
