//go:build go1.18

package fluentflag

import (
	"flag"
	"fmt"
	"os"
	"reflect"
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
			b := NewFlagBuilder[string](tt.name, tt.usage)
			if b.name != tt.name {
				t.Errorf("expected name %q, got %q", tt.name, b.name)
			}
			if b.usage != tt.usage {
				t.Errorf("expected usage %q, got %q", tt.usage, b.usage)
			}
		})
	}
}

func TestFlagBuilder_FluentAPI(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder[int]("num", "number flag").Alias('n').Default(42)
	if b.alias != 'n' {
		t.Errorf("expected alias 'n', got %v", b.alias)
	}
	if b.defaultVal != 42 {
		t.Errorf("expected default 42, got %v", b.defaultVal)
	}
}

func TestFlagBuilder_Build_Bool(t *testing.T) {
	resetFlags()
	var val bool
	b := NewFlagBuilder[bool]("flag", "bool flag").Default(true)
	b.Build(&val)
	args := []string{"--flag=false"}
	flag.CommandLine.Parse(args)
	if val != false {
		t.Errorf("expected false, got %v", val)
	}
}

func TestFlagBuilder_Build_Int(t *testing.T) {
	resetFlags()
	var val int
	b := NewFlagBuilder[int]("num", "int flag").Default(5)
	b.Build(&val)
	args := []string{"--num=99"}
	flag.CommandLine.Parse(args)
	if val != 99 {
		t.Errorf("expected 99, got %v", val)
	}
}

func TestFlagBuilder_Build_String_WithAlias(t *testing.T) {
	resetFlags()
	var val string
	b := NewFlagBuilder[string]("word", "string flag").Alias('w').Default("foo")
	b.Build(&val)
	args := []string{"-w", "bar"}
	flag.CommandLine.Parse(args)
	if val != "bar" {
		t.Errorf("expected 'bar', got %q", val)
	}
}

func TestFlagBuilder_BuildVar(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder[int64]("big", "big int").Default(123)
	ptr := b.BuildVar()
	args := []string{"--big=456"}
	flag.CommandLine.Parse(args)
	if *ptr != 456 {
		t.Errorf("expected 456, got %v", *ptr)
	}
}

func TestFlagBuilder_BuildSlice_String(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder[string]("item", "item flag")
	slice := b.BuildSlice()
	args := []string{"--item=foo", "--item=bar", "--item=baz"}
	flag.CommandLine.Parse(args)
	want := []string{"foo", "bar", "baz"}
	if !reflect.DeepEqual(*slice, want) {
		t.Errorf("expected %v, got %v", want, *slice)
	}
}

func TestFlagBuilder_BuildSlice_Int_WithAlias(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder[int]("num", "number").Alias('n')
	slice := b.BuildSlice()
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

// This test won't even compile, but we leave it here just for reference
// func TestParse_UnsupportedType(t *testing.T) {
// 	defer func() {
// 		if r := recover(); r == nil {
// 			t.Error("expected panic for unsupported type")
// 		}
// 	}()
// 	type mytype struct{}
// 	_ = NewFlagBuilder[mytype]("bad", "bad type")
// }

func TestFlagBuilder_Build_DefaultValue(t *testing.T) {
	resetFlags()
	var val uint
	b := NewFlagBuilder[uint]("count", "count flag").Default(7)
	b.Build(&val)
	flag.CommandLine.Parse([]string{})
	if val != 7 {
		t.Errorf("expected default 7, got %v", val)
	}
}

func TestFlagBuilder_BuildSlice_DefaultEmpty(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder[float64]("flt", "float flag")
	slice := b.BuildSlice()
	flag.CommandLine.Parse([]string{})
	if len(*slice) != 0 {
		t.Errorf("expected empty slice, got %v", *slice)
	}
}

func TestFlagBuilder_Build_Uint64(t *testing.T) {
	resetFlags()
	var val uint64
	b := NewFlagBuilder[uint64]("big", "big flag").Default(12345)
	b.Build(&val)
	args := []string{"--big=67890"}
	flag.CommandLine.Parse(args)
	if val != 67890 {
		t.Errorf("expected 67890, got %v", val)
	}
}

func TestFlagBuilder_UsageString(t *testing.T) {
	resetFlags()
	b := NewFlagBuilder[string]("foo", "foo usage")
	b.BuildVar()
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
	NewFlagBuilder[bool]("verbose", "enable verbose mode").Alias('v').Default(false).Build(&verbose)
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
			NewFlagBuilder[string]("string", "string flag").Alias('s').Default("default").Build(&strVal)
			NewFlagBuilder[bool]("bool", "bool flag").Alias('b').Default(false).Build(&boolVal)
			NewFlagBuilder[int]("int", "int flag").Alias('i').Build(&intVal)
			strSlice = NewFlagBuilder[string]("strslice", "string slice flag").Alias('S').BuildSlice()
			intSlice = NewFlagBuilder[int]("intslice", "int slice flag").Alias('I').BuildSlice()

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
