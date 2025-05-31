//go:build go1.18

package fluentflag

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func TestNewFlagBuilder_BasicFields(t *testing.T) {
	tests := []struct {
		name  string
		usage string
	}{
		{"testbool", "bool usage"},
		{"teststr", "string usage"},
		{"testint", "int usage"},
		{"testint64", "int64 usage"},
		{"testfloat64", "float64 usage"},
		{"testuint", "uint usage"},
		{"testuint64", "uint64 usage"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()
			b := NewFlagBuilder()
			var f any
			switch tt.name {
			case "testbool":
				f = b.BoolFlag(tt.name, tt.usage)
			case "teststr":
				f = b.StringFlag(tt.name, tt.usage)
			case "testint":
				f = b.IntFlag(tt.name, tt.usage)
			case "testint64":
				f = b.Int64Flag(tt.name, tt.usage)
			case "testfloat64":
				f = b.Float64Flag(tt.name, tt.usage)
			case "testuint":
				f = b.UintFlag(tt.name, tt.usage)
			case "testuint64":
				f = b.Uint64Flag(tt.name, tt.usage)
			}
			ff := f.(interface {
				GetName() string
				GetUsage() string
			})
			if ff.GetName() != tt.name {
				t.Errorf("expected name %q, got %q", tt.name, ff.GetName())
			}
			if ff.GetUsage() != tt.usage {
				t.Errorf("expected usage %q, got %q", tt.usage, ff.GetUsage())
			}
		})
	}
}

func (f *FluentFlag[T]) GetName() string  { return f.name }
func (f *FluentFlag[T]) GetUsage() string { return f.usage }

func TestFlagBuilder_FluentAPI(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder()
	f := b.IntFlag("num", "number flag").Alias('n').Default(42)
	if f.alias != 'n' {
		t.Errorf("expected alias 'n', got %v", f.alias)
	}
	if f.defaultVal != 42 {
		t.Errorf("expected default 42, got %v", f.defaultVal)
	}
}

func TestFlagBuilder_Build_Bool(t *testing.T) {
	resetFlags()
	var val bool
	b := NewFlagBuilder()
	b.BoolFlag("flag", "bool flag").Default(true).Build(&val)
	args := []string{"--flag=false"}
	flag.CommandLine.Parse(args)
	if val != false {
		t.Errorf("expected false, got %v", val)
	}
}

func TestFlagBuilder_Build_Int(t *testing.T) {
	resetFlags()
	var val int
	b := NewFlagBuilder()
	b.IntFlag("num", "int flag").Default(5).Build(&val)
	args := []string{"--num=99"}
	flag.CommandLine.Parse(args)
	if val != 99 {
		t.Errorf("expected 99, got %v", val)
	}
}

func TestFlagBuilder_Build_String_WithAlias(t *testing.T) {
	resetFlags()
	var val string
	b := NewFlagBuilder()
	b.StringFlag("word", "string flag").Alias('w').Default("foo").Build(&val)
	args := []string{"-w", "bar"}
	flag.CommandLine.Parse(args)
	if val != "bar" {
		t.Errorf("expected 'bar', got %q", val)
	}
}

func TestFlagBuilder_BuildVar(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder()
	ptr := b.Int64Flag("big", "big int").Default(123).BuildVar()
	args := []string{"--big=456"}
	flag.CommandLine.Parse(args)
	if *ptr != 456 {
		t.Errorf("expected 456, got %v", *ptr)
	}
}

func TestFlagBuilder_BuildSlice_String(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder()
	slice := b.StringFlag("item", "item flag").BuildSlice()
	args := []string{"--item=foo", "--item=bar", "--item=baz"}
	flag.CommandLine.Parse(args)
	want := []string{"foo", "bar", "baz"}
	if !reflect.DeepEqual(*slice, want) {
		t.Errorf("expected %v, got %v", want, *slice)
	}
}

func TestFlagBuilder_BuildSlice_Int_WithAlias(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder()
	slice := b.IntFlag("num", "number").Alias('n').BuildSlice()
	args := []string{"-n", "1", "-n", "2", "--num=3"}
	flag.CommandLine.Parse(args)
	want := []int{1, 2, 3}
	if !reflect.DeepEqual(*slice, want) {
		t.Errorf("expected %v, got %v", want, *slice)
	}
}

func TestParse_InvalidValue(t *testing.T) {
	_, err := parse[int]("notanint")
	if err == nil {
		t.Error("expected error for invalid int")
	}
	_, err = parse[bool]("notabool")
	if err == nil {
		t.Error("expected error for invalid bool")
	}
}

func TestNewFlagBuilderWithSet(t *testing.T) {
	resetFlags()
	customSet := flag.NewFlagSet("custom", flag.ContinueOnError)
	b := NewFlagBuilderWithSet(customSet)
	var val int
	b.IntFlag("num", "number").Default(1).Build(&val)
	customSet.Parse([]string{"--num=5"})
	if val != 5 {
		t.Errorf("expected 5, got %v", val)
	}
}

func TestBuildSlice_LongAndShortFlags(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder()
	slice := b.StringFlag("item", "item flag").Alias('i').BuildSlice()
	args := []string{"--item=foo", "-i", "bar"}
	flag.CommandLine.Parse(args)
	want := []string{"foo", "bar"}
	if !reflect.DeepEqual(*slice, want) {
		t.Errorf("expected %v, got %v", want, *slice)
	}
}

func TestAccumValuesString(t *testing.T) {
	var s []int
	acc := &accumValues[int]{target: &s}
	acc.Set("1")
	acc.Set("2")
	str := acc.String()
	if str != "[1 2]" {
		t.Errorf("expected '[1 2]', got %q", str)
	}
}

func TestBuildSliceDefaultAlwaysEmpty(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder()
	slice := b.StringFlag("s", "slice").Default("ignored").BuildSlice()
	flag.CommandLine.Parse([]string{})
	if len(*slice) != 0 {
		t.Errorf("expected empty slice, got %v", *slice)
	}
}

func TestAccumValuesSetErrorPropagation(t *testing.T) {
	var s []int
	acc := &accumValues[int]{target: &s}
	err := acc.Set("notanint")
	if err == nil {
		t.Error("expected error for invalid int")
	}
}

func TestAliasZeroNoShortFlag(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder()
	var val int
	b.IntFlag("num", "number").Alias(0).Default(1).Build(&val)
	args := []string{"-0", "2"}
	err := flag.CommandLine.Parse(args)
	if err == nil {
		t.Error("expected error for unknown shorthand")
	}
}

func TestFlagBuilder_InternalFields(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder()
	f := b.IntFlag("num", "number")
	if b.building != f {
		t.Error("expected building to be set to current flag")
	}
	var v int
	f.Build(&v)
	if b.building != nil {
		t.Error("expected building to be nil after Build")
	}
	if len(b.flagsBuilt) == 0 {
		t.Error("expected flagsBuilt to have at least one entry")
	}
}

func TestFlagBuilder_Build_DefaultValue(t *testing.T) {
	resetFlags()
	var val uint
	b := NewFlagBuilder()
	b.UintFlag("count", "count flag").Default(7).Build(&val)
	flag.CommandLine.Parse([]string{})
	if val != 7 {
		t.Errorf("expected default 7, got %v", val)
	}
}

func TestFlagBuilder_BuildSlice_DefaultEmpty(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder()
	slice := b.Float64Flag("flt", "float flag").BuildSlice()
	flag.CommandLine.Parse([]string{})
	if len(*slice) != 0 {
		t.Errorf("expected empty slice, got %v", *slice)
	}
}

func TestFlagBuilder_Build_Uint64(t *testing.T) {
	resetFlags()
	var val uint64
	b := NewFlagBuilder()
	b.Uint64Flag("big", "big flag").Default(12345).Build(&val)
	args := []string{"--big=67890"}
	flag.CommandLine.Parse(args)
	if val != 67890 {
		t.Errorf("expected 67890, got %v", val)
	}
}

func TestFlagBuilder_UsageString(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder()
	b.StringFlag("foo", "foo usage").BuildVar()
	fs := flag.CommandLine
	usage := ""
	fs.VisitAll(func(f *flag.Flag) {
		if f.Name == "foo" {
			usage = f.Usage
		}
	})
	if usage != "foo usage" {
		t.Errorf("expected usage 'foo usage', got %q", usage)
	}
}

func ExampleFlagBuilder() {
	resetFlags()
	var verbose bool
	NewFlagBuilder()
	NewFlagBuilder().BoolFlag("verbose", "enable verbose mode").Alias('v').Default(false).Build(&verbose)
	os.Args = []string{"cmd", "-v"}
	flag.CommandLine.Parse(os.Args[1:])
	fmt.Println(verbose)
	// Output: true
}

func TestFlagBuilder_TableDrivenCombos(t *testing.T) {
	type want struct {
		strVal   string
		boolVal  bool
		intVal   int
		strSlice []string
		intSlice []int
	}
	tests := []struct {
		name string
		args []string
		want want
	}{
		{
			name: "defaults",
			args: []string{},
			want: want{
				strVal:   "default",
				boolVal:  false,
				intVal:   0,
				strSlice: []string{},
				intSlice: []int{},
			},
		},
		{
			name: "all long flags",
			args: []string{"--string=foo", "--bool=true", "--int=42", "--strslice=one", "--strslice=two", "--intslice=1", "--intslice=2"},
			want: want{
				strVal:   "foo",
				boolVal:  true,
				intVal:   42,
				strSlice: []string{"one", "two"},
				intSlice: []int{1, 2},
			},
		},
		{
			name: "all short flags",
			args: []string{"-s", "bar", "-b", "-i", "7", "-S", "x", "-S", "y", "-I", "3", "-I", "4"},
			want: want{
				strVal:   "bar",
				boolVal:  true,
				intVal:   7,
				strSlice: []string{"x", "y"},
				intSlice: []int{3, 4},
			},
		},
		{
			name: "mixed flags",
			args: []string{"--string=hello", "-b", "--int=99", "-S", "a", "--strslice=b", "-I", "5"},
			want: want{
				strVal:   "hello",
				boolVal:  true,
				intVal:   99,
				strSlice: []string{"a", "b"},
				intSlice: []int{5},
			},
		},
		{
			name: "bool false explicit",
			args: []string{"--bool=false"},
			want: want{
				strVal:   "default",
				boolVal:  false,
				intVal:   0,
				strSlice: []string{},
				intSlice: []int{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()
			var (
				strVal   string
				boolVal  bool
				intVal   int
				strSlice *[]string
				intSlice *[]int
			)
			b := NewFlagBuilder()
			b.StringFlag("string", "string flag").Alias('s').Default("default").Build(&strVal)
			b.BoolFlag("bool", "bool flag").Alias('b').Default(false).Build(&boolVal)
			b.IntFlag("int", "int flag").Alias('i').Build(&intVal)
			strSlice = b.StringFlag("strslice", "string slice flag").Alias('S').BuildSlice()
			intSlice = b.IntFlag("intslice", "int slice flag").Alias('I').BuildSlice()

			err := flag.CommandLine.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if strVal != tt.want.strVal {
				t.Errorf("string: got %q, want %q", strVal, tt.want.strVal)
			}
			if boolVal != tt.want.boolVal {
				t.Errorf("bool: got %v, want %v", boolVal, tt.want.boolVal)
			}
			if intVal != tt.want.intVal {
				t.Errorf("int: got %v, want %v", intVal, tt.want.intVal)
			}
			if !reflect.DeepEqual(*strSlice, tt.want.strSlice) {
				t.Errorf("strSlice: got %v, want %v", *strSlice, tt.want.strSlice)
			}
			if !reflect.DeepEqual(*intSlice, tt.want.intSlice) {
				t.Errorf("intSlice: got %v, want %v", *intSlice, tt.want.intSlice)
			}
		})
	}
}

func TestFlagBuilder_PartiallyBuiltPanic(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for partially built flag, but did not panic")
		}
	}()
	b.BoolFlag("flag1", "usage1")
	b.IntFlag("flag2", "usage2") // should panic here
}

func TestMultipleBuildCalls(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder()
	f := b.IntFlag("num", "number")
	var v1, v2 int
	f.Build(&v1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for flag redefinition, but did not panic")
		} else if msg, ok := r.(error); ok && msg.Error() != "flag redefined: num" && !contains(msg.Error(), "flag redefined") {
			t.Errorf("unexpected panic: %v", r)
		}
	}()
	f.Build(&v2) // should panic
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (contains(s[1:], substr) || contains(s[:len(s)-1], substr))))
}

func TestFlagBuilder_UsageFormatting(t *testing.T) {
	resetFlags()
	builder := NewFlagBuilder()
	builder.StringFlag("name", "Command name for error messages").Alias('n').Default("foo").BuildVar()
	builder.BoolFlag("help", "Show this help message").Alias('h').BuildVar()
	builder.IntFlag("min-args", "Minimum number of non-option arguments").Alias('N').Default(-1).BuildVar()
	builder.IntFlag("max-args", "Maximum number of non-option arguments").Alias('X').Default(-1).BuildVar()
	builder.BoolFlag("ignore-unknown", "Ignore unknown options").Alias('i').BuildVar()
	builder.BoolFlag("stop-nonopt", "Stop scanning at first non-option").Alias('s').BuildVar()
	builder.BoolFlag("version", "Print version number").Alias('v').BuildVar()
	builder.IntFlag("exclusive", "Comma-separated mutually exclusive options").Alias('x').BuildVar()
	builder.StringFlag("this-is-a-very-long-flag-name-for-testing", "A very long flag name to test wrapping").Alias('L').Default("long").BuildVar()

	var buf strings.Builder
	builder.SetOutput(&buf)
	builder.PrintUsage()
	actual := strings.TrimRight(buf.String(), "\n")

	expected := `  -n, --name string        Command name for error messages (default "foo")
  -h, --help               Show this help message
  -N, --min-args int       Minimum number of non-option arguments (default -1)
  -X, --max-args int       Maximum number of non-option arguments (default -1)
  -i, --ignore-unknown     Ignore unknown options
  -s, --stop-nonopt        Stop scanning at first non-option
  -v, --version            Print version number
  -x, --exclusive int      Comma-separated mutually exclusive options
  -L, --this-is-a-very-long-flag-name-for-testing string
                           A very long flag name to test wrapping (default "long")`

	if actual != expected {
		t.Errorf("Usage output mismatch.\nGot:\n%s\nWant:\n%s", actual, expected)
	}
}
