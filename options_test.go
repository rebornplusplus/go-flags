package flags

import (
	"strings"
	"testing"
)

func TestPassDoubleDash(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`
	}{}

	p := NewParser(&opts, PassDoubleDash)
	ret, err := p.ParseArgs([]string{"-v", "--", "-v", "-g"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	assertStringArray(t, ret, []string{"-v", "-g"})
}

func TestPassAfterNonOption(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`
	}{}

	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs([]string{"-v", "arg", "-v", "-g"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	assertStringArray(t, ret, []string{"arg", "-v", "-g"})
}

func TestPassAfterNonOptionWithPositional(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []string `required:"yes"`
		} `positional-args:"yes"`
	}{}

	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs([]string{"-v", "arg", "-v", "-g"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	assertStringArray(t, ret, []string{})
	assertStringArray(t, opts.Positional.Rest, []string{"arg", "-v", "-g"})
}

func TestPassAfterNonOptionWithPositionalIntPass(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []int `required:"yes"`
		} `positional-args:"yes"`
	}{}

	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs([]string{"-v", "1", "2", "3"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	assertStringArray(t, ret, []string{})
	for i, rest := range opts.Positional.Rest {
		if rest != i+1 {
			assertErrorf(t, "Expected %v got %v", i+1, rest)
		}
	}
}

func TestPassAfterNonOptionWithPositionalIntFail(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []int `required:"yes"`
		} `positional-args:"yes"`
	}{}

	tests := []struct {
		opts        []string
		errContains string
		ret         []string
	}{
		{
			[]string{"-v", "notint1", "notint2", "notint3"},
			"notint1",
			[]string{"notint1", "notint2", "notint3"},
		},
		{
			[]string{"-v", "1", "notint2", "notint3"},
			"notint2",
			[]string{"1", "notint2", "notint3"},
		},
	}

	for _, test := range tests {
		p := NewParser(&opts, PassAfterNonOption)
		ret, err := p.ParseArgs(test.opts)

		if err == nil {
			assertErrorf(t, "Expected error")
			return
		}

		if !strings.Contains(err.Error(), test.errContains) {
			assertErrorf(t, "Expected the first illegal argument in the error")
		}

		assertStringArray(t, ret, test.ret)
	}
}

func TestTerminatedOptions(t *testing.T) {
	type testOpt struct {
		Args    []string `short:"a" long:"args" terminator:";"`
		Verbose bool     `short:"v"`
		Mode    int      `short:"m"`
		Value   int      `long:"val" terminator:"END"`
	}

	tests := []struct {
		parserOpts Options
		args       []string
		_Args      []string
		_Verbose   bool
		_Mode      int
		_Value     int
		ret        []string
		shouldErr  bool
	}{
		{
			args: []string{
				"--args", "bin", "-xyz", "--foo=bar", "--val", "5", "-v", "\"foo bar\"", ";",
				"-v",
				"--val", "10", "END",
				"--args=", "--no-delim",
			},
			_Args: []string{
				"bin -xyz --foo=bar --val 5 -v \"foo bar\"",
				"--no-delim",
			},
			_Verbose: true,
			_Value:   10,
		},
		{
			args:  []string{"--args", "--foo", "bar;", "-v"},
			_Args: []string{"--foo bar; -v"},
		},
		{
			args:  []string{"--args", "foo\tbar", "\"x y z\""},
			_Args: []string{"foo\tbar \"x y z\""},
		},
		{
			args:   []string{"--val", "1", "END", "--val", "2"},
			_Value: 2,
		},
		{
			parserOpts: PassDoubleDash,
			args:       []string{"--args", "--foo", "--", "bar", ";", "-m", "1", "--", "-v"},
			_Args:      []string{"--foo -- bar"},
			_Mode:      1,
			ret:        []string{"-v"},
		},
		{
			args:     []string{"-va", "--val", "5"},
			_Args:    []string{"--val 5"},
			_Verbose: true,
		},
		{
			args:      []string{"--val", "1", "2", "END", "-v"},
			shouldErr: true,
		},
	}

	for _, test := range tests {
		opts := testOpt{}
		p := NewParser(&opts, test.parserOpts)
		ret, err := p.ParseArgs(test.args)

		if err != nil {
			if !test.shouldErr {
				t.Fatalf("Unexpected error: %v", err)
			} else {
				continue
			}
		} else if test.shouldErr {
			t.Fatalf("Expected error")
		}

		if opts.Verbose != test._Verbose {
			t.Errorf("Expected Verbose to be %v, found %v", test._Verbose, opts.Verbose)
		}

		if opts.Mode != test._Mode {
			t.Errorf("Expected Mode to be %v, found %v", test._Mode, opts.Mode)
		}

		if opts.Value != test._Value {
			t.Errorf("Expected Value to be %v, found %v", test._Value, opts.Value)
		}

		assertStringArray(t, opts.Args, test._Args)

		assertStringArray(t, ret, test.ret)
	}
}
