package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestCommand_BasicFlags(t *testing.T) {
	var verbose, quiet bool
	var name string

	cmd := NewCommand("test", "test [options] [files]...")
	cmd.BoolVar(&verbose, 'v', "verbose", "verbose output")
	cmd.BoolVar(&quiet, 'q', "quiet", "quiet mode")
	cmd.StringVar(&name, 'n', "name", "target name")

	err := cmd.Parse([]string{"-v", "--name=foo", "file1.txt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Action() != ActionRun {
		t.Errorf("action = %v, want ActionRun", cmd.Action())
	}
	if !verbose || quiet {
		t.Errorf("verbose=%v, quiet=%v", verbose, quiet)
	}
	if name != "foo" {
		t.Errorf("name = %q, want 'foo'", name)
	}
	if !reflect.DeepEqual(cmd.Operands(), []string{"file1.txt"}) {
		t.Errorf("operands = %v, want ['file1.txt']", cmd.Operands())
	}
	if !reflect.DeepEqual(cmd.Args(), []string{"file1.txt"}) {
		t.Errorf("args = %v, want ['file1.txt']", cmd.Args())
	}
}

func TestCommand_ShortClusteringAndAttachedValue(t *testing.T) {
	var quiet, zero bool
	var lines string

	cmd := NewCommand("test", "test")
	cmd.BoolVar(&quiet, 'q', "quiet", "quiet mode")
	cmd.BoolVar(&zero, 'z', "zero-terminated", "null delim")
	cmd.StringVar(&lines, 'n', "lines", "lines count")

	// Cluster -qz followed by -n with attached negative value -5
	err := cmd.Parse([]string{"-qzn-5", "file.txt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !quiet || !zero {
		t.Errorf("quiet=%v, zero=%v, want true", quiet, zero)
	}
	if lines != "-5" {
		t.Errorf("lines = %q, want '-5'", lines)
	}
	if !reflect.DeepEqual(cmd.Operands(), []string{"file.txt"}) {
		t.Errorf("operands = %v", cmd.Operands())
	}
}

func TestCommand_NumericShorthand(t *testing.T) {
	var count string
	cmd := NewCommand("head", "head")
	cmd.EnableNumericShorthand("%s: invalid number of lines: '%s'", func(num string) error {
		count = num
		return nil
	})

	err := cmd.Parse([]string{"-25", "file.txt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != "25" {
		t.Errorf("count = %q, want '25'", count)
	}
	if !reflect.DeepEqual(cmd.Operands(), []string{"file.txt"}) {
		t.Errorf("operands = %v", cmd.Operands())
	}
}

func TestCommand_EndOfOptions(t *testing.T) {
	var verbose bool
	cmd := NewCommand("test", "test")
	cmd.BoolVar(&verbose, 'v', "verbose", "verbose")

	err := cmd.Parse([]string{"--", "-v", "arg"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verbose {
		t.Errorf("verbose should be false after '--'")
	}
	if !reflect.DeepEqual(cmd.Operands(), []string{"-v", "arg"}) {
		t.Errorf("operands = %v, want ['-v', 'arg']", cmd.Operands())
	}
}

func TestCommand_DefaultOperands(t *testing.T) {
	cmd := NewCommand("test", "test")
	cmd.SetDefaultOperands("-")

	err := cmd.Parse([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(cmd.Operands(), []string{"-"}) {
		t.Errorf("operands = %v, want ['-']", cmd.Operands())
	}
}

func TestCommand_HelpAndVersion(t *testing.T) {
	cmd := NewCommand("test", "test")

	err := cmd.Parse([]string{"--help"})
	if err != nil || cmd.Action() != ActionHelp {
		t.Errorf("expected ActionHelp, got %v (err: %v)", cmd.Action(), err)
	}

	cmd2 := NewCommand("test", "test")
	err = cmd2.Parse([]string{"--version"})
	if err != nil || cmd2.Action() != ActionVersion {
		t.Errorf("expected ActionVersion, got %v (err: %v)", cmd2.Action(), err)
	}
}

func TestCommand_Errors(t *testing.T) {
	cmd := NewCommand("mytool", "mytool")
	cmd.FlagFunc('f', "fail", "failing flag", true, func(val string) error {
		if val == "bad" {
			return errors.New("bad value")
		}
		return nil
	})

	// Unrecognized long
	err := cmd.Parse([]string{"--unknown"})
	if err == nil || err.Error() != "mytool: unrecognized option '--unknown'" {
		t.Errorf("got %v", err)
	}

	// Invalid short
	err = cmd.Parse([]string{"-x"})
	if err == nil || err.Error() != "mytool: invalid option -- 'x'" {
		t.Errorf("got %v", err)
	}

	// Missing short argument
	err = cmd.Parse([]string{"-f"})
	if err == nil || err.Error() != "mytool: option requires an argument -- 'f'" {
		t.Errorf("got %v", err)
	}

	// Missing long argument
	err = cmd.Parse([]string{"--fail"})
	if err == nil || err.Error() != "mytool: option '--fail' requires an argument" {
		t.Errorf("got %v", err)
	}

	// Custom handler error
	err = cmd.Parse([]string{"--fail=bad"})
	if err == nil || err.Error() != "bad value" {
		t.Errorf("got %v", err)
	}
}

func TestCommand_StopOnUnknown(t *testing.T) {
	var noNewline, escape bool
	newEchoCommand := func() *Command {
		cmd := NewCommandWithConfig("echo", EchoConfig(), "echo [OPTIONS]... [STRING]...")
		cmd.BoolVar(&noNewline, 'n', "", "no newline")
		cmd.BoolVar(&escape, 'e', "", "enable escape")
		return cmd
	}

	// 1. Regular flag group then operands
	cmd := newEchoCommand()
	err := cmd.Parse([]string{"-ne", "hello", "world"})
	if err != nil || !noNewline || !escape {
		t.Fatalf("unexpected: err=%v noNewline=%v escape=%v", err, noNewline, escape)
	}
	if !reflect.DeepEqual(cmd.Operands(), []string{"hello", "world"}) {
		t.Errorf("operands = %v, want ['hello', 'world']", cmd.Operands())
	}

	// 2. Cluster with unknown flag is not executed and passed as operand
	noNewline = false
	escape = false
	cmd = newEchoCommand()
	err = cmd.Parse([]string{"-nx", "hello"})
	if err != nil || noNewline {
		t.Fatalf("noNewline should remain false for invalid cluster -nx")
	}
	if !reflect.DeepEqual(cmd.Operands(), []string{"-nx", "hello"}) {
		t.Errorf("operands = %v, want ['-nx', 'hello']", cmd.Operands())
	}

	// 3. First non-flag stops option processing
	noNewline = false
	cmd = newEchoCommand()
	err = cmd.Parse([]string{"hello", "-n"})
	if err != nil || noNewline {
		t.Fatalf("noNewline should not be parsed after non-flag")
	}
	if !reflect.DeepEqual(cmd.Operands(), []string{"hello", "-n"}) {
		t.Errorf("operands = %v, want ['hello', '-n']", cmd.Operands())
	}

	// 4. Double dash is an operand, not swallowed
	cmd = newEchoCommand()
	err = cmd.Parse([]string{"--", "foo"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !reflect.DeepEqual(cmd.Operands(), []string{"--", "foo"}) {
		t.Errorf("operands = %v, want ['--', 'foo']", cmd.Operands())
	}

	// 5. Help at index 0 triggers ActionHelp, but after index 0 is an operand
	cmd = newEchoCommand()
	_ = cmd.Parse([]string{"-h"})
	if cmd.Action() != ActionHelp {
		t.Errorf("expected ActionHelp at index 0")
	}

	cmd = newEchoCommand()
	_ = cmd.Parse([]string{"-n", "-h"})
	if cmd.Action() == ActionHelp {
		t.Errorf("expected ActionRun when -h is not at index 0")
	}
	if !reflect.DeepEqual(cmd.Operands(), []string{"-h"}) {
		t.Errorf("operands = %v, want ['-h']", cmd.Operands())
	}
}

func TestCommand_Version(t *testing.T) {
	cmd := NewCommand("cat")
	if cmd.Version() != "cat (coreutils-go) 0.1.0" {
		t.Errorf("cmd.Version() = %q, want 'cat (coreutils-go) 0.1.0'", cmd.Version())
	}

	cmd.SetVersion("custom 9.9.9")
	if cmd.Version() != "custom 9.9.9" {
		t.Errorf("cmd.Version() = %q, want 'custom 9.9.9'", cmd.Version())
	}
}

func TestPureParser(t *testing.T) {
	fs := NewFlagSet()
	var flagVal bool
	fs.Add(&Flag{Short: 'a', Long: "all", Handler: func(_ string) error {
		flagVal = true
		return nil
	}})

	cfg := DefaultConfig()
	p := NewParser("puretool", fs, cfg)

	if err := p.Parse([]string{"-a", "operand1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !flagVal {
		t.Error("expected flagVal to be true")
	}
	if !reflect.DeepEqual(p.Operands(), []string{"operand1"}) {
		t.Errorf("operands = %v, want ['operand1']", p.Operands())
	}
}

func TestCommand_Execute(t *testing.T) {
	executed := false
	cmd := NewCommand("testexec", "testexec [OPTION]...")
	cmd.Run = func(c *Command, stdin io.Reader, stdout, stderr io.Writer) error {
		executed = true
		return nil
	}


	// 1. Successful execution
	err := cmd.Execute([]string{}, nil, nil, nil)
	if err != nil || !executed {
		t.Fatalf("expected err nil, got %v", err)
	}

	// 2. Help action handled
	var helpBuf bytes.Buffer
	err = cmd.Execute([]string{"--help"}, nil, &helpBuf, nil)
	if err != nil || !strings.Contains(helpBuf.String(), "Usage:") {
		t.Fatalf("expected err nil with help output, got %v: %q", err, helpBuf.String())
	}

	// 3. Error handling with custom OnError
	var errBuf bytes.Buffer
	cmd.OnError = func(c *Command, stderr io.Writer, err error) int {
		fmt.Fprintf(stderr, "custom: %v", err)
		return 99
	}
	err = cmd.Execute([]string{"--invalid"}, nil, nil, &errBuf)
	if err == nil || !strings.Contains(errBuf.String(), "custom:") {
		t.Fatalf("expected custom error, got %v: %q", err, errBuf.String())
	}

	// 4. Default execution using SetArgs without explicit arguments
	executed = false
	cmd.SetArgs([]string{})
	cmd.SetStreams(nil, &helpBuf, &errBuf)
	err = cmd.Execute()
	if err != nil || !executed {
		t.Fatalf("expected err nil without arguments, got %v", err)
	}
}



