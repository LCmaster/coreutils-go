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
		cmd := NewCommand("echo", "echo [OPTIONS]... [STRING]...")
		cmd.SetStopOnUnknown(true)
		cmd.EnableAutoShortVersion()
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

func TestCommand_HybridOptions(t *testing.T) {
	cmd := NewCommand("hybrid", "hybrid [OPTIONS]")

	// 1. Direct AddBool, AddString, AddInt with defaults
	cmd.AddBool('v', "verbose", "verbose output")
	cmd.AddBoolDefault('d', "dry-run", "dry run", true)
	cmd.AddString('o', "output", "output target")
	cmd.AddStringDefault('f', "format", "FMT", "output format", "text")
	cmd.AddInt('n', "count", "iteration count")
	cmd.AddIntDefault('p', "port", "PORT", "service port", 8080)

	// Verify defaults before parse
	if !cmd.Bool("dry-run") || !cmd.Bool("d") {
		t.Errorf("expected dry-run default true, got %v", cmd.Bool("dry-run"))
	}
	if cmd.String("format") != "text" || cmd.String("f") != "text" {
		t.Errorf("expected format default 'text', got %q", cmd.String("format"))
	}
	if cmd.Int("port") != 8080 || cmd.Int("p") != 8080 {
		t.Errorf("expected port default 8080, got %d", cmd.Int("port"))
	}

	// 2. Parse arguments mutating options
	err := cmd.Parse([]string{"-v", "-o", "report.json", "--format=json", "-n", "42", "--port=9090"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if !cmd.Bool("verbose") || !cmd.Bool("v") {
		t.Errorf("expected verbose=true, got %v", cmd.Bool("verbose"))
	}
	if cmd.String("output") != "report.json" || cmd.String("o") != "report.json" {
		t.Errorf("expected output='report.json', got %q", cmd.String("output"))
	}
	if cmd.String("format") != "json" || cmd.String("f") != "json" {
		t.Errorf("expected format='json', got %q", cmd.String("format"))
	}
	if cmd.Int("count") != 42 || cmd.Int("n") != 42 {
		t.Errorf("expected count=42, got %d", cmd.Int("count"))
	}
	if cmd.Int("port") != 9090 || cmd.Int("p") != 9090 {
		t.Errorf("expected port=9090, got %d", cmd.Int("port"))
	}

	// 3. HasBool, HasString, HasInt
	if !cmd.HasBool("verbose") || cmd.HasBool("non-existent") {
		t.Errorf("HasBool check failed")
	}
	if !cmd.HasString("format") || cmd.HasString("non-existent") {
		t.Errorf("HasString check failed")
	}
	if !cmd.HasInt("port") || cmd.HasInt("non-existent") {
		t.Errorf("HasInt check failed")
	}

	// 4. Manual Setters
	cmd.SetBool("custom-flag", true)
	if !cmd.Bool("custom-flag") || !cmd.Options.Bools["custom-flag"] {
		t.Errorf("SetBool failed")
	}

	cmd.SetString("custom-str", "hello")
	if cmd.String("custom-str") != "hello" || cmd.Options.Strings["custom-str"] != "hello" {
		t.Errorf("SetString failed")
	}

	cmd.SetInt("custom-int", 123)
	if cmd.Int("custom-int") != 123 || cmd.Options.Ints["custom-int"] != 123 {
		t.Errorf("SetInt failed")
	}

	// 5. Pointer-based BoolVar, StringVar, IntVar also tracking in Options
	var flagBool bool
	var flagStr string
	var flagInt int
	cmd2 := NewCommand("ptr-test")
	cmd2.BoolVar(&flagBool, 'b', "bool-flag", "flag")
	cmd2.StringVar(&flagStr, 's', "str-flag", "str")
	cmd2.IntVar(&flagInt, 'i', "int-flag", "int")

	err = cmd2.Parse([]string{"-b", "--str-flag=val", "-i", "100"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if !flagBool || !cmd2.Bool("bool-flag") || !cmd2.Bool("b") {
		t.Errorf("BoolVar hybrid sync failed")
	}
	if flagStr != "val" || cmd2.String("str-flag") != "val" || cmd2.String("s") != "val" {
		t.Errorf("StringVar hybrid sync failed")
	}
	if flagInt != 100 || cmd2.Int("int-flag") != 100 || cmd2.Int("i") != 100 {
		t.Errorf("IntVar hybrid sync failed")
	}

	// 6. Invalid integer argument returns error
	cmd3 := NewCommand("int-err")
	cmd3.AddInt('n', "num", "number")
	err = cmd3.Parse([]string{"-n", "not-a-number"})
	if err == nil {
		t.Errorf("expected error for invalid integer argument, got nil")
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

		// Trigger via short flag -q
		if err := cmd.Parse([]string{"-q"}); err != nil {
			t.Fatalf("unexpected error for -q: %v", err)
		}
		if quietCount != 1 {
			t.Errorf("expected quietCount=1, got %d", quietCount)
		}

		// Trigger via first alias --quiet
		if err := cmd.Parse([]string{"--quiet"}); err != nil {
			t.Fatalf("unexpected error for --quiet: %v", err)
		}
		if quietCount != 2 {
			t.Errorf("expected quietCount=2, got %d", quietCount)
		}

		// Trigger via second alias --silent
		if err := cmd.Parse([]string{"--silent"}); err != nil {
			t.Fatalf("unexpected error for --silent: %v", err)
		}
		if quietCount != 3 {
			t.Errorf("expected quietCount=3, got %d", quietCount)
		}
	})

	t.Run("FlagArgFunc with multiple long aliases and values", func(t *testing.T) {
		var receivedVal string
		cmd := NewCommand("head")
		cmd.FlagArgFunc('n', "lines, max-lines", "NUM", "line count", func(val string) error {
			receivedVal = val
			return nil
		})

		if err := cmd.Parse([]string{"--lines=42"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if receivedVal != "42" {
			t.Errorf("expected '42', got %q", receivedVal)
		}

		if err := cmd.Parse([]string{"--max-lines", "100"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if receivedVal != "100" {
			t.Errorf("expected '100', got %q", receivedVal)
		}
	})

	t.Run("AddBool with multiple aliases tracks options for all names", func(t *testing.T) {
		cmd := NewCommand("test")
		cmd.AddBool('q', "quiet, silent", "suppress output")

		if cmd.Bool("quiet") || cmd.Bool("silent") {
			t.Errorf("expected initial false for aliases")
		}

		if err := cmd.Parse([]string{"--silent"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !cmd.Bool("quiet") || !cmd.Bool("silent") || !cmd.Bool("q") {
			t.Errorf("expected all aliases to be true, got quiet=%v, silent=%v, q=%v",
				cmd.Bool("quiet"), cmd.Bool("silent"), cmd.Bool("q"))
		}
	})
}

