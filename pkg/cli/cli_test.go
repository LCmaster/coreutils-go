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
	cmd := NewCommand("test", "test [options] [files]...")
	cmd.FlagFunc('v', "verbose", "verbose output", false, func(_ string) error {
		cmd.SetBool("verbose", true)
		return nil
	})
	cmd.FlagFunc('q', "quiet", "quiet mode", false, func(_ string) error {
		cmd.SetBool("quiet", true)
		return nil
	})
	cmd.FlagFunc('n', "name", "target name", true, func(val string) error {
		cmd.SetString("name", val)
		return nil
	})

	err := cmd.Parse([]string{"-v", "--name=foo", "file1.txt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Action() != ActionRun {
		t.Errorf("action = %v, want ActionRun", cmd.Action())
	}
	if !cmd.Bool("verbose") || cmd.Bool("quiet") {
		t.Errorf("verbose=%v, quiet=%v", cmd.Bool("verbose"), cmd.Bool("quiet"))
	}
	if cmd.String("name") != "foo" {
		t.Errorf("name = %q, want 'foo'", cmd.String("name"))
	}
	if !reflect.DeepEqual(cmd.Operands(), []string{"file1.txt"}) {
		t.Errorf("operands = %v, want ['file1.txt']", cmd.Operands())
	}
}

func TestCommand_ShortClusteringAndAttachedValue(t *testing.T) {
	cmd := NewCommand("test", "test")
	cmd.FlagFunc('q', "quiet", "quiet mode", false, func(_ string) error {
		cmd.SetBool("quiet", true)
		return nil
	})
	cmd.FlagFunc('z', "zero-terminated", "null delim", false, func(_ string) error {
		cmd.SetBool("zero", true)
		return nil
	})
	cmd.FlagFunc('n', "lines", "lines count", true, func(val string) error {
		cmd.SetString("lines", val)
		return nil
	})

	// Cluster -qz followed by -n with attached negative value -5
	err := cmd.Parse([]string{"-qzn-5", "file.txt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cmd.Bool("quiet") || !cmd.Bool("zero") {
		t.Errorf("quiet=%v, zero=%v, want true", cmd.Bool("quiet"), cmd.Bool("zero"))
	}
	if cmd.String("lines") != "-5" {
		t.Errorf("lines = %q, want '-5'", cmd.String("lines"))
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

	// Numeric shorthand handler error formatted with NumericErrFormat
	cmdErr := NewCommand("head", "head")
	cmdErr.EnableNumericShorthand("%s: invalid number of lines: '%s'", func(_ string) error {
		return errors.New("out of range")
	})
	err = cmdErr.Parse([]string{"-0"})
	if err == nil || err.Error() != "head: invalid number of lines: '0'" {
		t.Errorf("expected custom error format, got %v", err)
	}
}

func TestCommand_EndOfOptions(t *testing.T) {
	cmd := NewCommand("test", "test")
	cmd.FlagFunc('v', "verbose", "verbose", false, func(_ string) error {
		cmd.SetBool("verbose", true)
		return nil
	})

	err := cmd.Parse([]string{"--", "-v", "arg"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Bool("verbose") {
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
	if len(cmd.ExplicitOperands()) != 0 || !cmd.UsedDefaultOperands() {
		t.Errorf("expected ExplicitOperands empty and UsedDefaultOperands true, got %v / %v", cmd.ExplicitOperands(), cmd.UsedDefaultOperands())
	}

	// Explicit "-" operand
	if err := cmd.Parse([]string{"-"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(cmd.ExplicitOperands(), []string{"-"}) || cmd.UsedDefaultOperands() {
		t.Errorf("expected ExplicitOperands=['-'] and UsedDefaultOperands=false, got %v / %v", cmd.ExplicitOperands(), cmd.UsedDefaultOperands())
	}
}

func TestCommand_OptionsResetOnRepeatedParse(t *testing.T) {
	cmd := NewCommand("reset-test")
	cmd.FlagFunc('v', "verbose", "verbose", false, func(_ string) error {
		cmd.SetBool("verbose", true)
		return nil
	})

	if err := cmd.Parse([]string{"-v"}); err != nil {
		t.Fatal(err)
	}
	if !cmd.Bool("verbose") {
		t.Fatalf("expected verbose=true after first parse")
	}

	// Second parse without flags should reset Options
	if err := cmd.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	if cmd.Bool("verbose") {
		t.Errorf("options not reset on second parse: verbose=%v", cmd.Bool("verbose"))
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
	newEchoCommand := func() *Command {
		cmd := NewCommand("echo", "echo [OPTIONS]... [STRING]...")
		cmd.SetStopOnUnknown(true)
		cmd.EnableAutoShortVersion()
		cmd.FlagFunc('n', "", "no newline", false, func(_ string) error {
			cmd.SetBool("noNewline", true)
			return nil
		})
		cmd.FlagFunc('e', "", "enable escape", false, func(_ string) error {
			cmd.SetBool("escape", true)
			return nil
		})
		return cmd
	}

	// 1. Regular flag group then operands
	cmd := newEchoCommand()
	err := cmd.Parse([]string{"-ne", "hello", "world"})
	if err != nil || !cmd.Bool("noNewline") || !cmd.Bool("escape") {
		t.Fatalf("unexpected: err=%v noNewline=%v escape=%v", err, cmd.Bool("noNewline"), cmd.Bool("escape"))
	}
	if !reflect.DeepEqual(cmd.Operands(), []string{"hello", "world"}) {
		t.Errorf("operands = %v, want ['hello', 'world']", cmd.Operands())
	}

	// 2. Cluster with unknown flag is not executed and passed as operand
	cmd = newEchoCommand()
	err = cmd.Parse([]string{"-nx", "hello"})
	if err != nil || cmd.Bool("noNewline") {
		t.Fatalf("noNewline should remain false for invalid cluster -nx")
	}
	if !reflect.DeepEqual(cmd.Operands(), []string{"-nx", "hello"}) {
		t.Errorf("operands = %v, want ['-nx', 'hello']", cmd.Operands())
	}

	// 3. First non-flag stops option processing
	cmd = newEchoCommand()
	err = cmd.Parse([]string{"hello", "-n"})
	if err != nil || cmd.Bool("noNewline") {
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
}

func TestCommand_OptionsSettersAndGetters(t *testing.T) {
	cmd := NewCommand("opts", "opts [OPTIONS]")

	cmd.SetBool("custom-flag", true)
	if !cmd.Bool("custom-flag") || !cmd.HasBool("custom-flag") || cmd.HasBool("missing") {
		t.Errorf("Bool / HasBool failed")
	}

	cmd.SetString("custom-str", "hello")
	if cmd.String("custom-str") != "hello" || !cmd.HasString("custom-str") || cmd.HasString("missing") {
		t.Errorf("String / HasString failed")
	}

	cmd.SetInt("custom-int", 123)
	if cmd.Int("custom-int") != 123 || !cmd.HasInt("custom-int") || cmd.HasInt("missing") {
		t.Errorf("Int / HasInt failed")
	}
}

func TestCommand_SoleArgumentHelpVersion(t *testing.T) {
	newCmd := func() *Command {
		cmd := NewCommand("true", "true [ignored command line arguments]", "true OPTION")
		cmd.EnableAutoShortHelp()
		cmd.EnableAutoShortVersion()
		return cmd
	}

	// 1. Sole --help
	cmd := newCmd()
	if err := cmd.Parse([]string{"--help"}); err != nil || cmd.Action() != ActionHelp {
		t.Errorf("expected ActionHelp for sole --help, got %v (err: %v)", cmd.Action(), err)
	}

	// 2. Sole -h
	cmd = newCmd()
	if err := cmd.Parse([]string{"-h"}); err != nil || cmd.Action() != ActionHelp {
		t.Errorf("expected ActionHelp for sole -h, got %v (err: %v)", cmd.Action(), err)
	}

	// 3. Sole --version
	cmd = newCmd()
	if err := cmd.Parse([]string{"--version"}); err != nil || cmd.Action() != ActionVersion {
		t.Errorf("expected ActionVersion for sole --version, got %v (err: %v)", cmd.Action(), err)
	}

	// 4. Sole -v
	cmd = newCmd()
	if err := cmd.Parse([]string{"-v"}); err != nil || cmd.Action() != ActionVersion {
		t.Errorf("expected ActionVersion for sole -v, got %v (err: %v)", cmd.Action(), err)
	}

	// 5. Multiple arguments with --help are treated as normal operands
	cmd = newCmd()
	if err := cmd.Parse([]string{"--help", "something"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Action() != ActionRun {
		t.Errorf("expected ActionRun when multiple args, got %v", cmd.Action())
	}
	if !reflect.DeepEqual(cmd.Operands(), []string{"--help", "something"}) {
		t.Errorf("operands = %v, want ['--help', 'something']", cmd.Operands())
	}

	// 6. Unknown flags treated as operands without errors
	cmd = newCmd()
	if err := cmd.Parse([]string{"-x", "--unknown"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Action() != ActionRun {
		t.Errorf("expected ActionRun, got %v", cmd.Action())
	}
	if !reflect.DeepEqual(cmd.Operands(), []string{"-x", "--unknown"}) {
		t.Errorf("operands = %v, want ['-x', '--unknown']", cmd.Operands())
	}

	// 7. Explicit EnableSoleArgumentHelpVersion without usage heuristic
	explicitCmd := NewCommand("testsole")
	explicitCmd.EnableSoleArgumentHelpVersion()
	if err := explicitCmd.Parse([]string{"--help"}); err != nil || explicitCmd.Action() != ActionHelp {
		t.Errorf("expected ActionHelp for explicit sole --help")
	}
	explicitCmd = NewCommand("testsole")
	explicitCmd.EnableSoleArgumentHelpVersion()
	if err := explicitCmd.Parse([]string{"--help", "operand"}); err != nil || explicitCmd.Action() != ActionRun {
		t.Errorf("expected ActionRun when not sole argument")
	}
}

func TestCommand_AutoShortHelpAndVersion(t *testing.T) {
	// DisableAutoShortHelp
	cmd := NewCommand("cat")
	cmd.DisableAutoShortHelp()
	err := cmd.Parse([]string{"-h"})
	if err == nil || !strings.Contains(err.Error(), "invalid option -- 'h'") {
		t.Errorf("expected invalid option -- 'h' when DisableAutoShortHelp, got: %v", err)
	}

	// Re-enable AutoShortHelp
	cmd = NewCommand("cat")
	cmd.DisableAutoShortHelp()
	cmd.EnableAutoShortHelp()
	err = cmd.Parse([]string{"-h"})
	if err != nil || cmd.Action() != ActionHelp {
		t.Errorf("expected ActionHelp with EnableAutoShortHelp, got action=%v err=%v", cmd.Action(), err)
	}

	// EnableAutoShortVersion
	cmd = NewCommand("echo")
	cmd.EnableAutoShortVersion()
	err = cmd.Parse([]string{"-v"})
	if err != nil || cmd.Action() != ActionVersion {
		t.Errorf("expected ActionVersion with EnableAutoShortVersion, got action=%v err=%v", cmd.Action(), err)
	}

	// DisableAutoShortVersion
	cmd = NewCommand("echo")
	cmd.EnableAutoShortVersion()
	cmd.DisableAutoShortVersion()
	err = cmd.Parse([]string{"-v"})
	if err == nil || !strings.Contains(err.Error(), "invalid option -- 'v'") {
		t.Errorf("expected invalid option -- 'v' when DisableAutoShortVersion, got: %v", err)
	}
}

func TestCommand_FlagAliases(t *testing.T) {
	t.Run("FlagFunc with multiple long aliases", func(t *testing.T) {
		quietCount := 0
		cmd := NewCommand("head")
		cmd.FlagFunc('q', "quiet, silent", "Never print headers", false, func(_ string) error {
			quietCount++
			return nil
		})

		if err := cmd.Parse([]string{"-q"}); err != nil || quietCount != 1 {
			t.Fatalf("unexpected: err=%v, quietCount=%d", err, quietCount)
		}
		if err := cmd.Parse([]string{"--quiet"}); err != nil || quietCount != 2 {
			t.Fatalf("unexpected: err=%v, quietCount=%d", err, quietCount)
		}
		if err := cmd.Parse([]string{"--silent"}); err != nil || quietCount != 3 {
			t.Fatalf("unexpected: err=%v, quietCount=%d", err, quietCount)
		}
	})

	t.Run("FlagArgFunc with multiple long aliases and values", func(t *testing.T) {
		var receivedVal string
		cmd := NewCommand("head")
		cmd.FlagArgFunc('n', "lines, max-lines", "NUM", "line count", func(val string) error {
			receivedVal = val
			return nil
		})

		if err := cmd.Parse([]string{"--lines=42"}); err != nil || receivedVal != "42" {
			t.Fatalf("unexpected: err=%v, val=%q", err, receivedVal)
		}
		if err := cmd.Parse([]string{"--max-lines", "100"}); err != nil || receivedVal != "100" {
			t.Fatalf("unexpected: err=%v, val=%q", err, receivedVal)
		}
	})
}
