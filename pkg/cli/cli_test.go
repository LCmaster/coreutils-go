package cli

import (
	"errors"
	"reflect"
	"testing"
)

func TestParser_BasicFlags(t *testing.T) {
	var verbose, quiet bool
	var name string

	p := NewParser("test", "test [options] [files]...")
	p.BoolVar(&verbose, 'v', "verbose", "verbose output")
	p.BoolVar(&quiet, 'q', "quiet", "quiet mode")
	p.StringVar(&name, 'n', "name", "target name")

	operands, action, err := p.Parse([]string{"-v", "--name=foo", "file1.txt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != ActionRun {
		t.Errorf("action = %v, want ActionRun", action)
	}
	if !verbose || quiet {
		t.Errorf("verbose=%v, quiet=%v", verbose, quiet)
	}
	if name != "foo" {
		t.Errorf("name = %q, want 'foo'", name)
	}
	if !reflect.DeepEqual(operands, []string{"file1.txt"}) {
		t.Errorf("operands = %v, want ['file1.txt']", operands)
	}
}

func TestParser_ShortClusteringAndAttachedValue(t *testing.T) {
	var quiet, zero bool
	var lines string

	p := NewParser("test", "test")
	p.BoolVar(&quiet, 'q', "quiet", "quiet mode")
	p.BoolVar(&zero, 'z', "zero-terminated", "null delim")
	p.StringVar(&lines, 'n', "lines", "lines count")

	// Cluster -qz followed by -n with attached negative value -5
	operands, _, err := p.Parse([]string{"-qzn-5", "file.txt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !quiet || !zero {
		t.Errorf("quiet=%v, zero=%v, want true", quiet, zero)
	}
	if lines != "-5" {
		t.Errorf("lines = %q, want '-5'", lines)
	}
	if !reflect.DeepEqual(operands, []string{"file.txt"}) {
		t.Errorf("operands = %v", operands)
	}
}

func TestParser_NumericShorthand(t *testing.T) {
	var count string
	p := NewParser("head", "head")
	p.EnableNumericShorthand("%s: invalid number of lines: '%s'", func(num string) error {
		count = num
		return nil
	})

	operands, _, err := p.Parse([]string{"-25", "file.txt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != "25" {
		t.Errorf("count = %q, want '25'", count)
	}
	if !reflect.DeepEqual(operands, []string{"file.txt"}) {
		t.Errorf("operands = %v", operands)
	}
}

func TestParser_EndOfOptions(t *testing.T) {
	var verbose bool
	p := NewParser("test", "test")
	p.BoolVar(&verbose, 'v', "verbose", "verbose")

	operands, _, err := p.Parse([]string{"--", "-v", "arg"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verbose {
		t.Errorf("verbose should be false after '--'")
	}
	if !reflect.DeepEqual(operands, []string{"-v", "arg"}) {
		t.Errorf("operands = %v, want ['-v', 'arg']", operands)
	}
}

func TestParser_DefaultOperands(t *testing.T) {
	p := NewParser("test", "test")
	p.SetDefaultOperands("-")

	operands, _, err := p.Parse([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(operands, []string{"-"}) {
		t.Errorf("operands = %v, want ['-']", operands)
	}
}

func TestParser_HelpAndVersion(t *testing.T) {
	p := NewParser("test", "test")

	_, actHelp, err := p.Parse([]string{"--help"})
	if err != nil || actHelp != ActionHelp {
		t.Errorf("expected ActionHelp, got %v (err: %v)", actHelp, err)
	}

	_, actVer, err := p.Parse([]string{"--version"})
	if err != nil || actVer != ActionVersion {
		t.Errorf("expected ActionVersion, got %v (err: %v)", actVer, err)
	}
}

func TestParser_Errors(t *testing.T) {
	p := NewParser("mytool", "mytool")
	p.FlagFunc('f', "fail", "failing flag", true, func(val string) error {
		if val == "bad" {
			return errors.New("bad value")
		}
		return nil
	})

	// Unrecognized long
	_, _, err := p.Parse([]string{"--unknown"})
	if err == nil || err.Error() != "mytool: unrecognized option '--unknown'" {
		t.Errorf("got %v", err)
	}

	// Invalid short
	_, _, err = p.Parse([]string{"-x"})
	if err == nil || err.Error() != "mytool: invalid option -- 'x'" {
		t.Errorf("got %v", err)
	}

	// Missing short argument
	_, _, err = p.Parse([]string{"-f"})
	if err == nil || err.Error() != "mytool: option requires an argument -- 'f'" {
		t.Errorf("got %v", err)
	}

	// Missing long argument
	_, _, err = p.Parse([]string{"--fail"})
	if err == nil || err.Error() != "mytool: option '--fail' requires an argument" {
		t.Errorf("got %v", err)
	}

	// Custom handler error
	_, _, err = p.Parse([]string{"--fail=bad"})
	if err == nil || err.Error() != "bad value" {
		t.Errorf("got %v", err)
	}
}
